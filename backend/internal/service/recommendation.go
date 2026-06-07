package service

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/gitsense/gitsense/backend/internal/model"
	"github.com/gitsense/gitsense/backend/internal/repository"
	"github.com/pgvector/pgvector-go"
)

// RecommendationService 处理推荐计算
// 多阶段 Pipeline: Embedding 召回 → Topic 扩展 → Language 过滤 → Ranking → Explanation
type RecommendationService struct {
	store               *repository.RepoStore
	cache               *repository.CacheStore
	strategyV1          SimilarityStrategy
	strategyV2          *WeightedSimilarityStrategyV2
	strategyV3          *WeightedSimilarityStrategyV3
	classifier          *EcosystemClassifier
	explanationGenerator *ExplanationGenerator
}

// NewRecommendationService 创建新的 RecommendationService
func NewRecommendationService(
	store *repository.RepoStore,
	cache *repository.CacheStore,
	strategyV1 SimilarityStrategy,
	strategyV2 *WeightedSimilarityStrategyV2,
	classifier *EcosystemClassifier,
) *RecommendationService {
	return &RecommendationService{
		store:                store,
		cache:                cache,
		strategyV1:           strategyV1,
		strategyV2:           strategyV2,
		classifier:           classifier,
		explanationGenerator: NewExplanationGenerator(),
	}
}

// SetStrategyV3 注入 V3 策略（由 main.go 调用）
func (s *RecommendationService) SetStrategyV3(v3 *WeightedSimilarityStrategyV3) {
	s.strategyV3 = v3
}

// GetRecommendations 获取相似仓库推荐（多阶段 Pipeline）
func (s *RecommendationService) GetRecommendations(ctx context.Context, fullName string, limit int, debug bool, strategyName string) ([]model.SimilarRepository, error) {
	if limit <= 0 {
		limit = 10
	}

	// 查缓存（非 debug 模式）
	if !debug {
		cached, err := s.cache.GetRecommendation(ctx, fullName, limit)
		if err == nil && cached != nil {
			return cached, nil
		}
	}

	// 查目标仓库
	target, err := s.store.GetByFullName(ctx, fullName)
	if err != nil {
		return nil, fmt.Errorf("get target repository: %w", err)
	}
	if target == nil {
		return nil, fmt.Errorf("repository not found: %s", fullName)
	}
	if target.EmbeddingStatus != model.EmbeddingDone {
		return nil, fmt.Errorf("repository embedding not ready: %s (status: %s)", fullName, target.EmbeddingStatus)
	}

	// --- Stage 1: Embedding 召回（主召回，top 50）---
	if target.Embedding == nil {
		return nil, fmt.Errorf("repository has no embedding: %s", fullName)
	}
	embeddingCandidates, err := s.store.SearchSimilar(ctx, fullName, *target.Embedding, 50)
	if err != nil {
		return nil, fmt.Errorf("embedding retrieval: %w", err)
	}
	log.Printf("[recommend] stage1 embedding recall: %d candidates for %s", len(embeddingCandidates), fullName)

	// --- Stage 2: Topic 扩展召回 ---
	topicCandidates, err := s.topicExpansionRecall(ctx, target)
	if err != nil {
		log.Printf("[recommend] stage2 topic recall error: %v", err)
	}
	log.Printf("[recommend] stage2 topic recall: %d additional candidates", len(topicCandidates))

	// 合并候选集（去重）
	candidateMap := make(map[string]model.Repository)
	for _, c := range embeddingCandidates {
		candidateMap[c.FullName] = c
	}
	for _, c := range topicCandidates {
		if _, exists := candidateMap[c.FullName]; !exists {
			candidateMap[c.FullName] = c
		}
	}

	// --- Stage 3: Language 过滤（降权，不过滤掉）---
	// 将候选转为列表
	var candidates []model.Repository
	for _, c := range candidateMap {
		candidates = append(candidates, c)
	}

	// --- Stage 4: Ranking ---
	var targetEmbeddingSlice []float32
	if target.Embedding != nil {
		targetEmbeddingSlice = target.Embedding.Slice()
	}
	targetFeatures := RepoFeatures{
		Embedding: targetEmbeddingSlice,
		Topics:    target.Topics,
		Language:  target.Language,
		Stars:     target.Stars,
	}

	// 选择策略
	useV2 := true
	useV3 := false
	if strategyName == "v1" {
		useV2 = false
	} else if strategyName == "v3" && s.strategyV3 != nil {
		useV2 = false
		useV3 = true
	}

	var results []model.SimilarRepository
	for _, candidate := range candidates {
		var candidateEmbeddingSlice []float32
		if candidate.Embedding != nil {
			candidateEmbeddingSlice = candidate.Embedding.Slice()
		}
		candidateFeatures := RepoFeatures{
			Embedding: candidateEmbeddingSlice,
			Topics:    candidate.Topics,
			Language:  candidate.Language,
			Stars:     candidate.Stars,
		}

		var score float64
		var reason RecommendationReason
		var breakdown *model.ScoreBreakdown

		if useV3 {
			if debug {
				bd, r, err := s.strategyV3.CalculateWithBreakdown(ctx, targetFeatures, candidateFeatures)
				if err != nil {
					continue
				}
				score = bd.FinalScore
				reason = r
				breakdown = &model.ScoreBreakdown{
					EmbeddingScore:  bd.EmbeddingScore,
					TopicScore:      bd.TopicScore,
					LanguageScore:   bd.LanguageScore,
					PopularityScore: bd.PopularityScore,
					TrendScore:      bd.TrendScore,
					FinalScore:      bd.FinalScore,
				}
			} else {
				sc, r, err := s.strategyV3.Calculate(ctx, targetFeatures, candidateFeatures)
				if err != nil {
					continue
				}
				score = sc
				reason = r
			}
		} else if useV2 {
			if debug {
				bd, r, err := s.strategyV2.CalculateWithBreakdown(ctx, targetFeatures, candidateFeatures)
				if err != nil {
					continue
				}
				score = bd.FinalScore
				reason = r
				breakdown = &model.ScoreBreakdown{
					EmbeddingScore:  bd.EmbeddingScore,
					TopicScore:      bd.TopicScore,
					LanguageScore:   bd.LanguageScore,
					PopularityScore: bd.PopularityScore,
					FinalScore:      bd.FinalScore,
				}
			} else {
				sc, r, err := s.strategyV2.Calculate(ctx, targetFeatures, candidateFeatures)
				if err != nil {
					continue
				}
				score = sc
				reason = r
			}
		} else {
			sc, r, err := s.strategyV1.Calculate(ctx, targetFeatures, candidateFeatures)
			if err != nil {
				continue
			}
			score = sc
			reason = r
		}

		// Language 不匹配降权 10%
		if target.Language != "" && candidate.Language != "" && target.Language != candidate.Language {
			score *= 0.9
		}

		// 生态分类
		ecoTag := s.classifier.Classify(candidate.Topics, candidate.Description)

		// 生态匹配
		targetEco := s.classifier.Classify(target.Topics, target.Description)
		ecoMatch := ecoTag != "" && ecoTag == targetEco

		// 构建 Features（Explanation Contract 的单一数据源）
		// reasons 必须完全基于此 features 派生
		features := model.RecommendationFeatures{
			EmbeddingScore:  reason.EmbeddingSimilarity,
			GraphScore:      0,
			TrendScore:      0,
			PopularityScore: reason.PopularityScore,
			TopicScore:      reason.TopicSimilarity,
		}
		if breakdown != nil {
			features.GraphScore = breakdown.EmbeddingScore * 0.8 // 近似 graph signal
			features.TrendScore = breakdown.TrendScore
		}

		// 生成 explanation reasons（纯派生，基于 features 单一数据源）
		reasons := s.explanationGenerator.Generate(features, ExplanationInput{
			Stars:          candidate.Stars,
			EcosystemMatch: ecoMatch,
			CommonTopics:   reason.CommonTopics,
		})

		results = append(results, model.SimilarRepository{
			Repository:   candidate,
			Score:        score,
			Features:     features,
			Reasons:      reasons,
			EcosystemTag: ecoTag,
			Breakdown:    breakdown,
		})
	}

	// --- Stage 5: 排序 + 截取 ---
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	// 写缓存（非 debug 模式）
	if !debug {
		_ = s.cache.SetRecommendation(ctx, fullName, limit, results)
	}

	return results, nil
}

// topicExpansionRecall Stage 2: Topic 扩展召回
// 使用分层阈值：>=3 topics 时 overlap>=2，否则 overlap>=1
func (s *RecommendationService) topicExpansionRecall(ctx context.Context, target *model.Repository) ([]model.Repository, error) {
	if len(target.Topics) == 0 {
		return nil, nil
	}

	// 查询有 topic 交集的仓库
	related, err := s.store.GetByTopics(ctx, target.Topics, 100)
	if err != nil {
		return nil, err
	}

	minOverlap := minTopicOverlap(len(target.Topics))

	var candidates []model.Repository
	for _, r := range related {
		// 排除自身
		if r.FullName == target.FullName {
			continue
		}
		// 只保留有 embedding 的
		if r.EmbeddingStatus != model.EmbeddingDone {
			continue
		}
		// 分层阈值过滤
		overlap := topicOverlapCount(target.Topics, r.Topics)
		if overlap >= minOverlap {
			candidates = append(candidates, r)
		}
	}

	return candidates, nil
}

// SearchByEmbedding 直接通过 embedding 向量搜索相似仓库
// 用于 /search/similar API（文本搜索，不依赖特定 repo）
func (s *RecommendationService) SearchByEmbedding(ctx context.Context, embedding pgvector.Vector, limit int) ([]model.Repository, error) {
	// 使用空 full_name 搜索（不排除任何仓库）
	return s.store.SearchSimilar(ctx, "", embedding, limit)
}

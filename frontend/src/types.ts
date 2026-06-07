export interface Repository {
  id: number;
  full_name: string;
  owner: string;
  name: string;
  description: string;
  language: string;
  stars: number;
  topics: string[];
}

export interface RecommendationFeatures {
  embedding_score: number;
  graph_score: number;
  trend_score: number;
  popularity_score: number;
  topic_score: number;
}

export interface SimilarRepository {
  repository: Repository;
  score: number;
  features: RecommendationFeatures;
  reasons: string[];
  ecosystem?: string;
}

export interface RecommendationResponse {
  similar_repositories: SimilarRepository[];
}

export interface TopicTrend {
  topic: string;
  growth_rate: number;
  trend_score: number;
  status: "rising" | "stable" | "declining";
  count_7d: number;
  count_prev_7d: number;
  window: string;
}

export interface TrendOverview {
  top_rising_topics: TopicTrend[];
  top_rising_ecosystems: Array<{
    ecosystem: string;
    growth_rate: number;
    trend_score: number;
    status: "rising" | "stable" | "declining";
  }>;
  window: string;
}

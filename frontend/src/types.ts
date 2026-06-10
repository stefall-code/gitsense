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

// --- Discovery Types ---

export interface StackRepo {
  full_name: string;
  description: string;
  stars: number;
  language: string;
  trend: string;
}

export interface Subcategory {
  name: string;
  repo_count: number;
  top_repos: StackRepo[];
  trending: StackRepo[];
}

export interface TechStackTree {
  ecosystem: string;
  categories: Subcategory[];
}

export interface EcosystemInfo {
  name: string;
  subcategory: string;
  repo_count: number;
  trend: string;
}

export interface RepoSummary {
  full_name: string;
  description: string;
  stars: number;
  language: string;
  topics: string[];
}

export interface DiscoveryResponse {
  repo: RepoSummary;
  ecosystem: EcosystemInfo;
  stack: TechStackTree;
  recommendations: SimilarRepository[];
}

export interface EcosystemSummary {
  name: string;
  repo_count: number;
  category_count: number;
  trend: string;
  trend_score: number;
}

export interface EcosystemsResponse {
  ecosystems: EcosystemSummary[];
}

export interface EcosystemDetail {
  name: string;
  repo_count: number;
  trend: string;
  trend_score: number;
  growth_rate: number;
  categories: Subcategory[];
  top_repos: StackRepo[];
}

export interface TrendingRepo {
  full_name: string;
  stars: number;
  language: string;
  trend: string;
  trend_score: number;
  subcategory: string;
}

export interface TrendingResponse {
  ecosystem: string;
  window: string;
  trending: TrendingRepo[];
}

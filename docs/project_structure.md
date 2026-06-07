# GitSense Project Structure

## Backend (`backend/`)

### Entry Point
- `cmd/server/main.go` — Application bootstrap, dependency injection, route registration

### Core Modules (`internal/`)

| Module | Path | Description |
|--------|------|-------------|
| **Bootstrap** | `bootstrap/` | BFS-based dataset collection from awesome-lists, topic search, and README link extraction. DB-backed queue with crash recovery. |
| **Config** | `config/` | Environment variable loading with sensible defaults. Structured config for Server, DB, Redis, GitHub, Embedding, Bootstrap. |
| **Ecosystem** | `ecosystem/` | Rule-based ecosystem classification. Structured `EcosystemRule{Name, Keywords, Topics}` for tech ecosystem detection. |
| **Embedding** | `embedding/` | `EmbeddingProvider` interface with 3 implementations: OpenAI, HTTP (Python sidecar), Mock. Supports single and batch generation. |
| **Graph** | `graph/` | Graph builder (repo edges + topic edges + ecosystem map), store (PostgreSQL), service (neighborhood queries, path finding, explanation). |
| **Handler** | `handler/` | HTTP handlers for all API endpoints. Gin context processing, request validation, response formatting. |
| **Middleware** | `middleware/` | `AdminAuth` middleware — Bearer token verification for `/admin/*` routes. |
| **Model** | `model/` | Data structures: Repository, RecommendationResult, SimilarRepository, RecommendationFeatures. |
| **Repository** | `repository/` | PostgreSQL data access layer for repositories. CRUD, embedding update, search. |
| **Router** | `router/` | Gin route registration. Public API (`/api/v1/*`) and admin API (`/admin/*`) groups. |
| **Search** | `search/` | Semantic search service combining embedding similarity with metadata filtering. |
| **Service** | `service/` | Core recommendation engine. SimilarityStrategy interface with V1/V2/V3 implementations. ExplanationGenerator. EmbeddingWorker. |
| **Trend** | `trend/` | Trend detection using 7-day growth rate with Laplace smoothing and tanh compression. Redis-cached with worker refresh. |
| **Audit** | `audit/` | Graph quality audit and ranking ablation analysis. Threshold sweep, coverage metrics, A/B testing. |
| **GitHub** | `github/` | GitHub API client with dual rate limiter (Core + Search API). Repository, README, topics, language fetching. |

### Database Migrations (`migrations/`)

| File | Description |
|------|-------------|
| `001_init_repositories.sql` | Repositories table with pgvector embedding column |
| `002_graph_tables.sql` | repo_edges, topic_edges, ecosystem_map tables |
| `003_resize_embedding.sql` | Embedding dimension migration (provider switch) |
| `004_trend_tables.sql` | Trend cache and metadata tables |
| `005_explanation_tables.sql` | Explanation and features storage |
| `006_bootstrap_tables.sql` | bootstrap_jobs and bootstrap_queue tables |

## Frontend (`frontend/`)

- `src/pages/SearchPage.tsx` — Main search page with repo input
- `src/pages/ResultPage.tsx` — Recommendation results with RepoCards
- `src/components/RepoCard.tsx` — Repository card with score, reasons, ecosystem
- `src/api/api.ts` — API client for backend communication
- `src/types/types.ts` — TypeScript type definitions

## Embedding Service (`embedding-service/`)

- `main.py` — FastAPI service with `/embed` and `/embed/batch` endpoints
- Uses `sentence-transformers/all-MiniLM-L6-v2` (384 dimensions)
- Model downloaded at Docker build time, loaded at runtime
- Supports batch inference for efficient processing

## Key Design Decisions

1. **Embedding Provider Interface** — Swappable between local MiniLM and OpenAI, with dimension-aware migration
2. **Similarity Strategy Interface** — V1 (emb+topic+lang), V2 (+popularity), V3 (+graph+trend), coexisting with `?strategy=` switch
3. **DB-backed Bootstrap Queue** — Persistent queue with crash recovery, not Redis (durability priority)
4. **Dual Rate Limiter** — Separate Core and Search API buckets matching GitHub's actual rate limit structure
5. **Explanation as Separate Layer** — Ranking produces scores, explanation produces reasons, single data source (RecommendationFeatures)

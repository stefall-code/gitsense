# GitSense System Metrics — Release v1.0

## Dataset Statistics

| Metric | Value |
|--------|-------|
| **Repositories** | 5,972 |
| **Embeddings** | 5,972 / 5,972 (100% success) |
| **Unique Topics** | 14,874 |
| **Topic Edges** | 6,326 |
| **Ecosystems** | 10 |
| **Repo Edges** | 794 |
| **Bootstrap Success Rate** | 69.6% |

## Graph Metrics

| Metric | Value |
|--------|-------|
| Active Nodes (repos with edges) | 558 |
| Graph Coverage | 9.34% |
| Average Degree | 1.42 |
| Connected Components | 159 |
| Largest Component | ~20 nodes |

## Embedding Quality

- **Model**: all-MiniLM-L6-v2 (384 dimensions)
- **Average cosine similarity (top-10 neighbors)**: 0.537
- **Max cosine similarity observed**: 0.854
- **Embedding failure rate**: 0%

## Recommendation Quality

### Top-10 Recommendation for langchain-ai/langgraph

| Rank | Repository | Score | Key Reason |
|------|-----------|-------|------------|
| 1 | langchain-ai/langchain | 0.815 | High embedding similarity + shared topics |
| 2 | openai/openai-agents-python | 0.611 | Strong embedding match |
| 3 | langflow-ai/langflow | 0.587 | Related LLM framework |
| 4 | HKUDS/nanobot | 0.551 | AI Agent ecosystem |
| 5 | ag2ai/ag2 | 0.530 | Multi-agent framework |

### A/B Test: V3 (Full) vs V1 (Embedding Only)

- **Average Top-10 Overlap**: 77.1%
- **Average Graph Discovery**: 2.3 new repos per query
- **Max Graph Discovery**: 5 (apache/kafka)

## Trend Detection

- **Method**: 7-day growth rate with Laplace smoothing + tanh compression
- **Formula**: `growth_rate = (count_7d + 1) / (count_prev_7d + 1)`, `trend_score = tanh(log(growth_rate))`
- **Cache**: Redis with 12h TTL, worker refresh every 6h

## Bootstrap System

- **Seed Sources**: 10 awesome-lists + 10 topic queries
- **BFS Depth**: Max 2 (awesome → README → topic expansion)
- **Quality Filter**: stars >= 100, forks >= 10, language != null, updated within 2 years
- **Workers**: 3 concurrent with global GitHub rate limiter
- **Queue**: PostgreSQL-backed with UNIQUE(repo_full_name) deduplication

## Known Limitations

1. **Graph Coverage Low** — Only 9.34% of repos have repo_edges, limiting graph signal in ranking
2. **MiniLM Similarity Range** — 384-dim model produces compressed similarity range (avg 0.537), higher-dim models would improve discrimination
3. **No User System** — No authentication, favorites, or personalization
4. **No Online Demo** — Requires local Docker deployment

## Validation Methodology

- Phase 9.1: 50 repos — verified embedding + recommendation pipeline
- Phase 9.2: 200 repos — verified no duplicates, stable embeddings
- Phase 9.3: 5,000+ repos — verified at scale, measured all metrics
- Phase 9.5: Graph audit — threshold sweep, ablation study
- Phase 9.6: Graph rebuild — threshold 0.75 → 0.60, edges 6 → 794

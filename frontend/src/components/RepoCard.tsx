import type { SimilarRepository } from "../types";

interface RepoCardProps {
  item: SimilarRepository;
  rank: number;
}

function scoreColor(score: number): string {
  if (score >= 0.8) return "#16a34a";
  if (score >= 0.6) return "#2563eb";
  if (score >= 0.4) return "#d97706";
  return "#6b7280";
}

export default function RepoCard({ item, rank }: RepoCardProps) {
  const repo = item.repository;

  return (
    <div
      style={{
        border: "1px solid #e5e7eb",
        borderRadius: 8,
        padding: 16,
        marginBottom: 12,
        background: "#fff",
      }}
    >
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
        <div style={{ flex: 1 }}>
          <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 4 }}>
            <span
              style={{
                background: "#f3f4f6",
                borderRadius: 4,
                padding: "2px 8px",
                fontSize: 12,
                fontWeight: 600,
                color: "#374151",
              }}
            >
              #{rank}
            </span>
            <a
              href={`https://github.com/${repo.full_name}`}
              target="_blank"
              rel="noopener noreferrer"
              style={{ fontSize: 16, fontWeight: 600, color: "#2563eb", textDecoration: "none" }}
            >
              {repo.full_name}
            </a>
            {repo.language && (
              <span style={{ fontSize: 12, color: "#6b7280", background: "#f9fafb", padding: "1px 6px", borderRadius: 4 }}>
                {repo.language}
              </span>
            )}
            {item.ecosystem && (
              <span style={{ fontSize: 11, color: "#7c3aed", background: "#f5f3ff", padding: "1px 6px", borderRadius: 4 }}>
                {item.ecosystem}
              </span>
            )}
          </div>
          <p style={{ margin: "4px 0 8px", color: "#4b5563", fontSize: 14, lineHeight: 1.5 }}>
            {repo.description || "No description"}
          </p>
        </div>
        <div
          style={{
            background: scoreColor(item.score),
            color: "#fff",
            borderRadius: 6,
            padding: "4px 10px",
            fontSize: 14,
            fontWeight: 700,
            whiteSpace: "nowrap",
            marginLeft: 12,
          }}
        >
          {item.score.toFixed(2)}
        </div>
      </div>

      {/* Reasons */}
      {item.reasons && item.reasons.length > 0 && (
        <div style={{ marginTop: 8, borderTop: "1px solid #f3f4f6", paddingTop: 8 }}>
          <ul style={{ margin: 0, paddingLeft: 16, color: "#6b7280", fontSize: 13 }}>
            {item.reasons.map((reason, i) => (
              <li key={i} style={{ marginBottom: 2 }}>
                {reason}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Topics */}
      {repo.topics && repo.topics.length > 0 && (
        <div style={{ marginTop: 8, display: "flex", flexWrap: "wrap", gap: 4 }}>
          {repo.topics.slice(0, 6).map((topic) => (
            <span
              key={topic}
              style={{
                fontSize: 11,
                background: "#eff6ff",
                color: "#1d4ed8",
                padding: "1px 6px",
                borderRadius: 4,
              }}
            >
              {topic}
            </span>
          ))}
        </div>
      )}

      {/* Features bar (compact) */}
      <div style={{ marginTop: 8, display: "flex", gap: 12, fontSize: 11, color: "#9ca3af" }}>
        {item.features && (
          <>
            {item.features.embedding_score > 0 && (
              <span>emb: {item.features.embedding_score.toFixed(2)}</span>
            )}
            {item.features.graph_score > 0 && (
              <span>graph: {item.features.graph_score.toFixed(2)}</span>
            )}
            {item.features.trend_score > 0 && (
              <span>trend: {item.features.trend_score.toFixed(2)}</span>
            )}
            {item.features.topic_score > 0 && (
              <span>topic: {item.features.topic_score.toFixed(2)}</span>
            )}
            {item.features.popularity_score > 0 && (
              <span>pop: {item.features.popularity_score.toFixed(2)}</span>
            )}
          </>
        )}
      </div>
    </div>
  );
}

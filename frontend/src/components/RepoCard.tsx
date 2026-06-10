import type { SimilarRepository } from "../types";
import { trackRecommendationClick } from "../analytics";
import { useI18n } from "../i18n";

interface RepoCardProps {
  item: SimilarRepository;
  rank: number;
  fromRepo?: string;
}

function scoreColor(score: number): string {
  if (score >= 0.8) return "#16a34a";
  if (score >= 0.6) return "#2563eb";
  if (score >= 0.4) return "#d97706";
  return "#6b7280";
}

export default function RepoCard({ item, rank, fromRepo }: RepoCardProps) {
  const repo = item.repository;
  const { tr } = useI18n();

  const handleClick = () => {
    trackRecommendationClick(repo.full_name, fromRepo || "", rank);
  };

  return (
    <div
      style={{
        border: "1px solid #e5e7eb",
        borderRadius: 10,
        padding: 16,
        background: "#fff",
        boxShadow: "0 1px 3px rgba(0,0,0,0.04)",
        transition: "box-shadow 0.2s, border-color 0.2s",
      }}
      onMouseEnter={(e) => { e.currentTarget.style.boxShadow = "0 4px 12px rgba(0,0,0,0.08)"; e.currentTarget.style.borderColor = "#c7d2fe"; }}
      onMouseLeave={(e) => { e.currentTarget.style.boxShadow = "0 1px 3px rgba(0,0,0,0.04)"; e.currentTarget.style.borderColor = "#e5e7eb"; }}
    >
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start" }}>
        <div style={{ flex: 1 }}>
          <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 4, flexWrap: "wrap" }}>
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
              onClick={handleClick}
              style={{ fontSize: 16, fontWeight: 600, color: "#2563eb", textDecoration: "none", overflow: "hidden", textOverflow: "ellipsis" }}
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
            borderRadius: 8,
            padding: "4px 10px",
            fontSize: 14,
            fontWeight: 700,
            whiteSpace: "nowrap",
            marginLeft: 12,
            boxShadow: "0 1px 2px rgba(0,0,0,0.1)",
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
                {tr(reason)}
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

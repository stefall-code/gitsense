import type { StackRepo } from "../types";

interface Props {
  repos: StackRepo[];
  ecosystemName: string;
}

export default function TrendPanel({ repos, ecosystemName }: Props) {
  if (repos.length === 0) return null;

  return (
    <div style={{ marginBottom: 20 }}>
      <h3 style={{ fontSize: 16, fontWeight: 700, color: "#111827", marginBottom: 12 }}>
        🔥 Trending in {ecosystemName}
      </h3>
      <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
        {repos.map((repo) => (
          <span
            key={repo.full_name}
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: 4,
              background: "#f0fdf4",
              border: "1px solid #bbf7d0",
              borderRadius: 16,
              padding: "4px 12px",
              fontSize: 13,
              color: "#166534",
              fontWeight: 500,
            }}
          >
            ↑ {repo.full_name.split("/")[1]}
            <span style={{ fontSize: 11, color: "#6b7280" }}>★{repo.stars.toLocaleString()}</span>
          </span>
        ))}
      </div>
    </div>
  );
}

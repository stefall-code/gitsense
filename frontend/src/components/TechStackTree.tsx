import { useState } from "react";
import type { TechStackTree, StackRepo } from "../types";

interface Props {
  stack: TechStackTree;
}

function RepoItem({ repo }: { repo: StackRepo }) {
  const trendColor =
    repo.trend === "rising"
      ? "#16a34a"
      : repo.trend === "declining"
      ? "#dc2626"
      : "#6b7280";

  const trendIcon =
    repo.trend === "rising" ? "↑" : repo.trend === "declining" ? "↓" : "→";

  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "6px 0 6px 20px",
        fontSize: 13,
      }}
    >
      <div style={{ display: "flex", alignItems: "center", gap: 8, minWidth: 0 }}>
        <span style={{ color: "#374151", fontWeight: 500, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
          {repo.full_name}
        </span>
        {repo.language && (
          <span style={{ fontSize: 11, color: "#9ca3af" }}>{repo.language}</span>
        )}
      </div>
      <div style={{ display: "flex", alignItems: "center", gap: 8, flexShrink: 0 }}>
        <span style={{ fontSize: 12, color: "#6b7280" }}>★ {repo.stars.toLocaleString()}</span>
        <span style={{ fontSize: 11, color: trendColor, fontWeight: 600 }}>
          {trendIcon}
        </span>
      </div>
    </div>
  );
}

export default function TechStackTree({ stack }: Props) {
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});

  const toggle = (name: string) => {
    setExpanded((prev) => ({ ...prev, [name]: !prev[name] }));
  };

  return (
    <div style={{ marginBottom: 20 }}>
      <h3 style={{ fontSize: 16, fontWeight: 700, color: "#111827", marginBottom: 12 }}>
        Tech Stack
      </h3>

      {stack.categories.map((cat) => {
        const isOpen = expanded[cat.name] ?? true;
        return (
          <div key={cat.name} style={{ marginBottom: 4 }}>
            <button
              onClick={() => toggle(cat.name)}
              style={{
                display: "flex",
                alignItems: "center",
                gap: 6,
                width: "100%",
                background: "none",
                border: "none",
                cursor: "pointer",
                padding: "12px 4px",
                fontSize: 14,
                fontWeight: 600,
                color: "#374151",
                textAlign: "left",
              }}
            >
              <span style={{ fontSize: 10, color: "#9ca3af", transition: "transform 0.15s", transform: isOpen ? "rotate(90deg)" : "rotate(0deg)", display: "inline-block" }}>
                ▶
              </span>
              {cat.name}
              <span style={{ fontSize: 12, color: "#9ca3af", fontWeight: 400 }}>
                ({cat.repo_count})
              </span>
            </button>

            {isOpen && (
              <div>
                {cat.top_repos.map((repo) => (
                  <RepoItem key={repo.full_name} repo={repo} />
                ))}
                {cat.trending.length > 0 && (
                  <div style={{ paddingLeft: 20, marginTop: 4 }}>
                    <span style={{ fontSize: 11, color: "#16a34a", fontWeight: 600 }}>
                      🔥 Trending:
                    </span>{" "}
                    {cat.trending.map((r) => r.full_name.split("/")[1]).join(", ")}
                  </div>
                )}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}

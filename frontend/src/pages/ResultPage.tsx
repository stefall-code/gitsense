import { useEffect, useState } from "react";
import { useSearchParams, useNavigate } from "react-router-dom";
import { getRecommendations } from "../api";
import type { RecommendationResponse } from "../types";
import RepoCard from "../components/RepoCard";

export default function ResultPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const owner = searchParams.get("owner") || "";
  const repo = searchParams.get("repo") || "";

  const [data, setData] = useState<RecommendationResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!owner || !repo) {
      navigate("/");
      return;
    }

    setLoading(true);
    setError("");

    getRecommendations(owner, repo, { limit: 10 })
      .then(setData)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [owner, repo, navigate]);

  return (
    <div style={{ maxWidth: 720, margin: "0 auto", padding: "32px 20px" }}>
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <button
          onClick={() => navigate("/")}
          style={{
            background: "none",
            border: "none",
            color: "#2563eb",
            cursor: "pointer",
            fontSize: 14,
            padding: 0,
            marginBottom: 12,
          }}
        >
          &larr; New Search
        </button>

        <h1 style={{ fontSize: 24, fontWeight: 700, color: "#111827", marginBottom: 4 }}>
          {owner}/{repo}
        </h1>
      </div>

      {/* Loading */}
      {loading && (
        <div style={{ textAlign: "center", padding: 48, color: "#9ca3af" }}>
          Loading recommendations...
        </div>
      )}

      {/* Error */}
      {error && (
        <div style={{ padding: 16, background: "#fef2f2", borderRadius: 8, color: "#dc2626", fontSize: 14 }}>
          {error}
        </div>
      )}

      {/* Results */}
      {data && (
        <>
          <div style={{ marginBottom: 16, color: "#6b7280", fontSize: 14 }}>
            Found <strong>{data.similar_repositories.length}</strong> similar repositories
          </div>

          {data.similar_repositories.length === 0 && (
            <div style={{ textAlign: "center", padding: 48, color: "#9ca3af" }}>
              No similar repositories found
            </div>
          )}

          {data.similar_repositories.map((item, i) => (
            <RepoCard key={item.repository.full_name} item={item} rank={i + 1} />
          ))}
        </>
      )}
    </div>
  );
}

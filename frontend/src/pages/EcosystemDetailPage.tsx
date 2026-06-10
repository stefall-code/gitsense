import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { getEcosystem } from "../api";
import type { EcosystemDetail as EcosystemDetailType, StackRepo } from "../types";
import TechStackTree from "../components/TechStackTree";
import { useI18n } from "../i18n";

function RepoItem({ repo, onClick }: { repo: StackRepo; onClick: () => void }) {
  const trendColor =
    repo.trend === "rising" ? "#16a34a" : repo.trend === "declining" ? "#dc2626" : "#6b7280";
  const trendBg =
    repo.trend === "rising" ? "#dcfce7" : repo.trend === "declining" ? "#fee2e2" : "#f3f4f6";
  const trendIcon = repo.trend === "rising" ? "↑" : repo.trend === "declining" ? "↓" : "→";

  return (
    <div
      onClick={onClick}
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "10px 16px",
        background: "#fff",
        border: "1px solid #e5e7eb",
        borderRadius: 8,
        marginBottom: 8,
        cursor: "pointer",
        transition: "border-color 0.2s, box-shadow 0.2s",
      }}
      onMouseEnter={(e) => { e.currentTarget.style.borderColor = "#93c5fd"; e.currentTarget.style.boxShadow = "0 2px 8px rgba(0,0,0,0.06)"; }}
      onMouseLeave={(e) => { e.currentTarget.style.borderColor = "#e5e7eb"; e.currentTarget.style.boxShadow = "none"; }}
    >
      <div style={{ minWidth: 0 }}>
        <div style={{ fontSize: 14, fontWeight: 600, color: "#111827", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
          {repo.full_name}
        </div>
        {repo.description && (
          <div style={{ fontSize: 12, color: "#6b7280", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap", marginTop: 2 }}>
            {repo.description}
          </div>
        )}
      </div>
      <div style={{ display: "flex", alignItems: "center", gap: 10, flexShrink: 0 }}>
        {repo.language && (
          <span style={{ fontSize: 11, color: "#9ca3af" }}>{repo.language}</span>
        )}
        <span style={{ fontSize: 12, color: "#6b7280" }}>★ {repo.stars.toLocaleString()}</span>
        <span style={{ fontSize: 12, color: trendColor, fontWeight: 600, background: trendBg, padding: "1px 6px", borderRadius: 6 }}>{trendIcon}</span>
      </div>
    </div>
  );
}

export default function EcosystemDetailPage() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const { t } = useI18n();
  const [data, setData] = useState<EcosystemDetailType | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!name) {
      navigate("/ecosystems");
      return;
    }

    setLoading(true);
    getEcosystem(name)
      .then(setData)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [name, navigate]);

  if (loading) {
    return (
      <div style={{ maxWidth: 720, margin: "0 auto", padding: "32px 20px", textAlign: "center", color: "#9ca3af" }}>
        {t("ecoDetail.loading")}
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ maxWidth: 720, margin: "0 auto", padding: "32px 20px" }}>
        <div style={{ padding: 16, background: "#fef2f2", borderRadius: 8, color: "#dc2626", fontSize: 14 }}>
          {error}
        </div>
      </div>
    );
  }

  if (!data) return null;

  const trendColor =
    data.trend === "rising" ? "#16a34a" : data.trend === "declining" ? "#dc2626" : "#6b7280";
  const trendBg =
    data.trend === "rising" ? "#dcfce7" : data.trend === "declining" ? "#fee2e2" : "#f3f4f6";
  const trendIcon = data.trend === "rising" ? "↑" : data.trend === "declining" ? "↓" : "→";
  const trendLabel = data.trend === "rising" ? t("trend.rising") : data.trend === "declining" ? t("trend.declining") : t("trend.stable");

  return (
    <div style={{ maxWidth: 720, margin: "0 auto", padding: "32px 20px" }}>
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <div style={{ display: "flex", alignItems: "center", gap: 10, marginBottom: 8 }}>
          <h1 style={{ fontSize: 24, fontWeight: 700, color: "#1e40af", margin: 0 }}>
            {data.name}
          </h1>
          <span
            style={{
              fontSize: 12,
              fontWeight: 600,
              color: trendColor,
              background: trendBg,
              padding: "4px 10px",
              borderRadius: 10,
            }}
          >
            {trendIcon} {trendLabel}
          </span>
        </div>
        <div style={{ fontSize: 14, color: "#6b7280" }}>
          {data.repo_count} {t("ecoList.repos")} · {data.categories.length} {t("ecoList.categories")}
        </div>
      </div>

      {/* Overview */}
      <div
        style={{
          background: "#f9fafb",
          border: "1px solid #e5e7eb",
          borderRadius: 10,
          padding: "16px 20px",
          marginBottom: 24,
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(120px, 1fr))",
          gap: 16,
        }}
      >
        <div>
          <div style={{ fontSize: 12, color: "#9ca3af", marginBottom: 4 }}>{t("ecoDetail.trendScore")}</div>
          <div style={{ fontSize: 20, fontWeight: 700, color: "#111827" }}>
            {data.trend_score.toFixed(2)}
          </div>
        </div>
        <div>
          <div style={{ fontSize: 12, color: "#9ca3af", marginBottom: 4 }}>{t("ecoDetail.growthRate")}</div>
          <div style={{ fontSize: 20, fontWeight: 700, color: "#111827" }}>
            {data.growth_rate.toFixed(2)}x
          </div>
        </div>
        <div>
          <div style={{ fontSize: 12, color: "#9ca3af", marginBottom: 4 }}>{t("ecoDetail.categories")}</div>
          <div style={{ fontSize: 20, fontWeight: 700, color: "#111827" }}>
            {data.categories.length}
          </div>
        </div>
      </div>

      {/* Technology Stack */}
      <TechStackTree stack={{ ecosystem: data.name, categories: data.categories }} />

      {/* Top Repos */}
      {data.top_repos.length > 0 && (
        <div style={{ marginTop: 24 }}>
          <h3 style={{ fontSize: 16, fontWeight: 700, color: "#111827", marginBottom: 12 }}>
            {t("ecoDetail.topRepos")}
          </h3>
          {data.top_repos.map((repo) => {
            const [owner, name] = repo.full_name.split("/");
            return <RepoItem key={repo.full_name} repo={repo} onClick={() => navigate(`/r/${encodeURIComponent(owner)}/${encodeURIComponent(name)}`)} />;
          })}
        </div>
      )}
    </div>
  );
}

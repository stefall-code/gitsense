import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { getDiscovery } from "../api";
import type { DiscoveryResponse } from "../types";
import RepoCard from "../components/RepoCard";
import EcosystemHeader from "../components/EcosystemHeader";
import TechStackTree from "../components/TechStackTree";
import TrendPanel from "../components/TrendPanel";
import { trackShare, trackRecommendationClick, trackEcosystemOpen } from "../analytics";
import { useI18n } from "../i18n";

export default function RepoProfilePage() {
  const { owner, repo } = useParams<{ owner: string; repo: string }>();
  const navigate = useNavigate();
  const { t } = useI18n();
  const [data, setData] = useState<DiscoveryResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [shareMsg, setShareMsg] = useState("");

  useEffect(() => {
    if (!owner || !repo) {
      navigate("/");
      return;
    }

    setLoading(true);
    setError("");

    getDiscovery(owner, repo, { limit: 10 })
      .then(setData)
      .catch((err) => {
        if (err.message.includes("not found") || err.message.includes("404")) {
          setError("NOT_FOUND");
        } else if (err.message.includes("not indexed") || err.message.includes("no embedding")) {
          setError("NOT_INDEXED");
        } else {
          setError(err.message);
        }
      })
      .finally(() => setLoading(false));
  }, [owner, repo, navigate]);

  const handleShare = async () => {
    const url = window.location.href;
    try {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(url);
      } else {
        const textarea = document.createElement("textarea");
        textarea.value = url;
        textarea.style.position = "fixed";
        textarea.style.opacity = "0";
        document.body.appendChild(textarea);
        textarea.select();
        document.execCommand("copy");
        document.body.removeChild(textarea);
      }
      setShareMsg(t("repo.share.copied"));
      trackShare(`${owner}/${repo}`);
      setTimeout(() => setShareMsg(""), 2000);
    } catch {
      setShareMsg(t("repo.share.failed"));
    }
  };

  const trendingRepos = data?.stack.categories.flatMap((c) => c.trending) || [];

  if (loading) {
    return (
      <div style={{ maxWidth: 720, margin: "0 auto", padding: "32px 20px" }}>
        <div style={{ textAlign: "center", padding: 80, color: "#9ca3af", fontSize: 16 }}>
          {t("repo.loading")}
        </div>
      </div>
    );
  }

  if (error === "NOT_FOUND") {
    return (
      <div style={{ maxWidth: 720, margin: "0 auto", padding: "32px 20px" }}>
        <div style={{ textAlign: "center", padding: "60px 20px" }}>
          <div style={{ fontSize: 48, marginBottom: 16 }}>🔍</div>
          <h2 style={{ fontSize: 20, fontWeight: 600, color: "#111827", marginBottom: 8 }}>
            {t("repo.notFound.title")}
          </h2>
          <p style={{ color: "#6b7280", fontSize: 14, marginBottom: 24 }}>
            <strong>{owner}/{repo}</strong> {t("repo.notFound.desc")}
          </p>
          <button
            onClick={() => navigate("/")}
            style={{ padding: "10px 24px", background: "#2563eb", color: "#fff", border: "none", borderRadius: 8, cursor: "pointer", fontSize: 14, fontWeight: 600 }}
          >
            {t("repo.notFound.button")}
          </button>
        </div>
      </div>
    );
  }

  if (error === "NOT_INDEXED") {
    return (
      <div style={{ maxWidth: 720, margin: "0 auto", padding: "32px 20px" }}>
        <div style={{ textAlign: "center", padding: "60px 20px" }}>
          <div style={{ fontSize: 48, marginBottom: 16 }}>📋</div>
          <h2 style={{ fontSize: 20, fontWeight: 600, color: "#111827", marginBottom: 8 }}>
            {t("repo.notIndexed.title")}
          </h2>
          <p style={{ color: "#6b7280", fontSize: 14, marginBottom: 24 }}>
            <strong>{owner}/{repo}</strong> {t("repo.notIndexed.desc")}
          </p>
          <div style={{ display: "flex", gap: 12, justifyContent: "center" }}>
            <button
              onClick={() => navigate("/")}
              style={{ padding: "10px 24px", background: "#2563eb", color: "#fff", border: "none", borderRadius: 8, cursor: "pointer", fontSize: 14, fontWeight: 600 }}
            >
              {t("repo.notIndexed.explore")}
            </button>
            <a
              href={`https://github.com/${owner}/${repo}`}
              target="_blank"
              rel="noopener noreferrer"
              style={{ padding: "10px 24px", background: "#f3f4f6", color: "#374151", border: "none", borderRadius: 8, cursor: "pointer", fontSize: 14, fontWeight: 600, textDecoration: "none" }}
            >
              {t("repo.notIndexed.github")}
            </a>
          </div>
        </div>
      </div>
    );
  }

  if (error && error !== "NOT_FOUND" && error !== "NOT_INDEXED") {
    return (
      <div style={{ maxWidth: 720, margin: "0 auto", padding: "32px 20px" }}>
        <div style={{ textAlign: "center", padding: "60px 20px" }}>
          <div style={{ fontSize: 48, marginBottom: 16 }}>⚠️</div>
          <h2 style={{ fontSize: 20, fontWeight: 600, color: "#111827", marginBottom: 8 }}>
            {t("repo.error.title")}
          </h2>
          <p style={{ color: "#6b7280", fontSize: 14, marginBottom: 24 }}>
            {t("repo.error.desc")}
          </p>
          <button
            onClick={() => window.location.reload()}
            style={{ padding: "10px 24px", background: "#2563eb", color: "#fff", border: "none", borderRadius: 8, cursor: "pointer", fontSize: 14, fontWeight: 600 }}
          >
            {t("repo.error.retry")}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: 720, margin: "0 auto", padding: "32px 20px" }}>
      {/* Navigation */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 24 }}>
        <button
          onClick={() => navigate("/")}
          style={{ background: "none", border: "none", color: "#2563eb", cursor: "pointer", fontSize: 14, padding: "10px 0" }}
        >
          &larr; {t("repo.newSearch")}
        </button>
        <button
          onClick={handleShare}
          style={{ padding: "10px 16px", background: "#f3f4f6", border: "1px solid #e5e7eb", borderRadius: 6, cursor: "pointer", fontSize: 13, color: "#374151" }}
        >
          {shareMsg || t("repo.share")}
        </button>
      </div>

      {/* Repo Header */}
      <div style={{ marginBottom: 24 }}>
        <h1 style={{ fontSize: 24, fontWeight: 700, color: "#111827", marginBottom: 4, display: "flex", alignItems: "center", gap: 8 }}>
          <a
            href={`https://github.com/${owner}/${repo}`}
            target="_blank"
            rel="noopener noreferrer"
            style={{ color: "#111827", textDecoration: "none" }}
          >
            {owner}/{repo}
          </a>
          <span style={{ fontSize: 14 }}>↗</span>
        </h1>
        {data?.repo.description && (
          <p style={{ fontSize: 14, color: "#6b7280", margin: "4px 0 8px" }}>{data.repo.description}</p>
        )}
        <div style={{ display: "flex", gap: 12, fontSize: 13, color: "#6b7280", flexWrap: "wrap" }}>
          {data?.repo.stars > 0 && <span>⭐ {data.repo.stars.toLocaleString()}</span>}
          {data?.repo.language && <span>🔤 {data.repo.language}</span>}
          {data?.repo.topics && data.repo.topics.length > 0 && (
            <span>{data.repo.topics.slice(0, 5).map(t2 => (
              <span key={t2} style={{ background: "#eff6ff", color: "#2563eb", padding: "2px 8px", borderRadius: 4, fontSize: 12, marginRight: 4 }}>
                {t2}
              </span>
            ))}</span>
          )}
        </div>
      </div>

      {/* Ecosystem Header */}
      {data?.ecosystem && data.ecosystem.name && (
        <div style={{ marginTop: 4 }}>
          <EcosystemHeader ecosystem={data.ecosystem} fromRepo={`${owner}/${repo}`} />
        </div>
      )}

      {/* No Ecosystem */}
      {data && (!data.ecosystem || !data.ecosystem.name) && (
        <div style={{ padding: 16, background: "#f9fafb", borderRadius: 8, marginBottom: 24, fontSize: 14, color: "#6b7280", border: "1px solid #e5e7eb" }}>
          {t("repo.noEcosystem")}
        </div>
      )}

      {/* Separator */}
      <div style={{ borderTop: "1px solid #e5e7eb", margin: "24px 0" }} />

      {/* Recommendations */}
      <div style={{ marginBottom: 16, color: "#6b7280", fontSize: 14 }}>
        {t("repo.similar.prefix", "")}Found <strong>{data?.recommendations.length || 0}</strong> {t("repo.similar")}
      </div>

      {data && data.recommendations.length === 0 && (
        <div style={{ textAlign: "center", padding: 48, color: "#9ca3af", fontSize: 14 }}>
          <div style={{ fontSize: 32, marginBottom: 8 }}>🔎</div>
          <p>{t("repo.noSimilar")}</p>
          <p style={{ fontSize: 13 }}>{t("repo.noSimilar.hint")}</p>
        </div>
      )}

      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        {data?.recommendations.map((item, i) => (
          <RepoCard key={item.repository.full_name} item={item} rank={i + 1} fromRepo={`${owner}/${repo}`} />
        ))}
      </div>

      {/* Tech Stack Tree */}
      {data?.stack && data.stack.categories.length > 0 && (
        <div style={{ marginTop: 32 }}>
          <TechStackTree stack={data.stack} />
        </div>
      )}

      {/* Trend Panel */}
      {trendingRepos.length > 0 && data?.ecosystem && (
        <TrendPanel repos={trendingRepos} ecosystemName={data.ecosystem.name} />
      )}
    </div>
  );
}

import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { getEcosystems } from "../api";
import type { EcosystemSummary } from "../types";
import { useI18n } from "../i18n";

export default function EcosystemsListPage() {
  const navigate = useNavigate();
  const { t } = useI18n();
  const [ecosystems, setEcosystems] = useState<EcosystemSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    getEcosystems()
      .then((res) => setEcosystems(res.ecosystems))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div style={{ maxWidth: 720, margin: "0 auto", padding: "32px 20px" }}>
      <h1 style={{ fontSize: 24, fontWeight: 700, color: "#111827", marginBottom: 24 }}>
        {t("ecoList.title")}
      </h1>

      {loading && (
        <div style={{ textAlign: "center", padding: 48, color: "#9ca3af" }}>
          {t("ecoList.loading")}
        </div>
      )}

      {error && (
        <div style={{ padding: 16, background: "#fef2f2", borderRadius: 8, color: "#dc2626", fontSize: 14 }}>
          {error}
        </div>
      )}

      <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
        {ecosystems.map((eco) => {
          const trendColor =
            eco.trend === "rising" ? "#16a34a" : eco.trend === "declining" ? "#dc2626" : "#6b7280";
          const trendBg =
            eco.trend === "rising" ? "#dcfce7" : eco.trend === "declining" ? "#fee2e2" : "#f3f4f6";
          const trendIcon = eco.trend === "rising" ? "↑" : eco.trend === "declining" ? "↓" : "→";
          const trendLabel = eco.trend === "rising" ? t("trend.rising") : eco.trend === "declining" ? t("trend.declining") : t("trend.stable");

          return (
            <div
              key={eco.name}
              onClick={() => navigate(`/ecosystem/${encodeURIComponent(eco.name)}`)}
              style={{
                background: "#fff",
                border: "1px solid #e5e7eb",
                borderRadius: 10,
                padding: "16px 20px",
                cursor: "pointer",
                boxShadow: "0 1px 3px rgba(0,0,0,0.04)",
                transition: "border-color 0.2s, box-shadow 0.2s, transform 0.2s",
              }}
              onMouseEnter={(e) => { e.currentTarget.style.borderColor = "#93c5fd"; e.currentTarget.style.boxShadow = "0 4px 12px rgba(0,0,0,0.08)"; e.currentTarget.style.transform = "translateY(-1px)"; }}
              onMouseLeave={(e) => { e.currentTarget.style.borderColor = "#e5e7eb"; e.currentTarget.style.boxShadow = "0 1px 3px rgba(0,0,0,0.04)"; e.currentTarget.style.transform = "translateY(0)"; }}
            >
              <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
                <div>
                  <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 4 }}>
                    <span style={{ fontSize: 16, fontWeight: 600, color: "#1e40af" }}>
                      {eco.name}
                    </span>
                    <span
                      style={{
                        fontSize: 11,
                        fontWeight: 600,
                        color: trendColor,
                        background: trendBg,
                        padding: "2px 6px",
                        borderRadius: 8,
                      }}
                    >
                      {trendIcon} {trendLabel}
                    </span>
                  </div>
                  <div style={{ fontSize: 13, color: "#6b7280" }}>
                    {eco.repo_count} {t("ecoList.repos")} · {eco.category_count} {t("ecoList.categories")}
                  </div>
                </div>
                <div style={{ fontSize: 12, color: "#9ca3af" }}>
                  score: {eco.trend_score.toFixed(2)}
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

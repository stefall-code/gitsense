import type { EcosystemInfo } from "../types";
import { useNavigate } from "react-router-dom";
import { trackEcosystemOpen } from "../analytics";

interface Props {
  ecosystem: EcosystemInfo;
  fromRepo?: string;
}

export default function EcosystemHeader({ ecosystem, fromRepo }: Props) {
  const navigate = useNavigate();

  const handleClick = () => {
    trackEcosystemOpen(ecosystem.name, fromRepo || "");
    navigate(`/ecosystem/${encodeURIComponent(ecosystem.name)}`);
  };

  const trendColor =
    ecosystem.trend === "rising"
      ? "#16a34a"
      : ecosystem.trend === "declining"
      ? "#dc2626"
      : "#6b7280";

  const trendBg =
    ecosystem.trend === "rising"
      ? "#dcfce7"
      : ecosystem.trend === "declining"
      ? "#fee2e2"
      : "#f3f4f6";

  const trendIcon =
    ecosystem.trend === "rising"
      ? "↑"
      : ecosystem.trend === "declining"
      ? "↓"
      : "→";

  return (
    <div
      style={{
        background: "linear-gradient(135deg, #eff6ff 0%, #f0fdf4 100%)",
        border: "1px solid #bfdbfe",
        borderRadius: 12,
        padding: "16px 20px",
        marginBottom: 20,
        cursor: "pointer",
        boxShadow: "0 1px 3px rgba(0,0,0,0.05)",
        transition: "box-shadow 0.2s, border-color 0.2s",
      }}
      onClick={handleClick}
      onMouseEnter={(e) => { e.currentTarget.style.boxShadow = "0 4px 12px rgba(0,0,0,0.1)"; e.currentTarget.style.borderColor = "#93c5fd"; }}
      onMouseLeave={(e) => { e.currentTarget.style.boxShadow = "0 1px 3px rgba(0,0,0,0.05)"; e.currentTarget.style.borderColor = "#bfdbfe"; }}
    >
      <div style={{ display: "flex", alignItems: "center", gap: 10, marginBottom: 6 }}>
        <span style={{ fontSize: 20 }}>🏗️</span>
        <span style={{ fontSize: 18, fontWeight: 700, color: "#1e40af" }}>
          {ecosystem.name}
        </span>
        <span
          style={{
            fontSize: 12,
            fontWeight: 600,
            color: trendColor,
            background: trendBg,
            padding: "2px 8px",
            borderRadius: 10,
          }}
        >
          {trendIcon} {ecosystem.trend}
        </span>
      </div>
      <div style={{ fontSize: 13, color: "#6b7280" }}>
        Subcategory: <strong>{ecosystem.subcategory}</strong> · {ecosystem.repo_count} repos in ecosystem
      </div>
    </div>
  );
}

import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { parseRepoInput } from "../api";
import { trackSearch } from "../analytics";
import { useI18n } from "../i18n";

export default function SearchPage() {
  const [input, setInput] = useState("");
  const [error, setError] = useState("");
  const navigate = useNavigate();
  const { t } = useI18n();

  const DEMOS = [
    {
      title: t("search.demo.ai.title"),
      repo: "langchain-ai/langchain",
      description: t("search.demo.ai.desc"),
      icon: "🤖",
    },
    {
      title: t("search.demo.go.title"),
      repo: "gin-gonic/gin",
      description: t("search.demo.go.desc"),
      icon: "⚡",
    },
    {
      title: t("search.demo.data.title"),
      repo: "pandas-dev/pandas",
      description: t("search.demo.data.desc"),
      icon: "📊",
    },
  ];

  const handleSearch = () => {
    setError("");
    const parsed = parseRepoInput(input);
    if (!parsed) {
      setError(t("search.error"));
      return;
    }
    trackSearch(`${parsed.owner}/${parsed.repo}`);
    navigate(`/r/${encodeURIComponent(parsed.owner)}/${encodeURIComponent(parsed.repo)}`);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") handleSearch();
  };

  const handleDemo = (repo: string) => {
    trackSearch(repo);
    const [owner, name] = repo.split("/");
    navigate(`/r/${encodeURIComponent(owner)}/${encodeURIComponent(name)}`);
  };

  return (
    <div style={{ minHeight: "calc(100vh - 49px)", display: "flex", flexDirection: "column", background: "linear-gradient(180deg, #f0f5ff 0%, #ffffff 40%)" }}>
      {/* Hero */}
      <div style={{
        flex: 1,
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        padding: "60px 20px 40px",
        maxWidth: 800,
        margin: "0 auto",
        width: "100%",
      }}>
        <h1 style={{
          fontSize: 52,
          fontWeight: 800,
          color: "#111827",
          marginBottom: 12,
          textAlign: "center",
          lineHeight: 1.15,
          letterSpacing: "-0.02em",
        }}>
          {t("search.title")}
        </h1>
        <p style={{
          fontSize: 20,
          color: "#4b5563",
          textAlign: "center",
          marginBottom: 4,
          maxWidth: 560,
          lineHeight: 1.5,
        }}>
          {t("search.subtitle1")}
        </p>
        <p style={{
          fontSize: 20,
          color: "#2563eb",
          fontWeight: 600,
          textAlign: "center",
          marginBottom: 40,
          maxWidth: 560,
        }}>
          {t("search.subtitle2")}
        </p>

        {/* Search Box */}
        <div style={{
          display: "flex",
          gap: 8,
          width: "100%",
          maxWidth: 560,
          marginBottom: 8,
        }}>
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={t("search.placeholder")}
            style={{
              flex: 1,
              padding: "14px 18px",
              fontSize: 16,
              border: "2px solid #e5e7eb",
              borderRadius: 10,
              outline: "none",
              background: "#fff",
              boxShadow: "0 1px 3px rgba(0,0,0,0.06)",
              transition: "border-color 0.2s, box-shadow 0.2s",
            }}
            onFocus={(e) => { e.target.style.borderColor = "#2563eb"; e.target.style.boxShadow = "0 0 0 3px rgba(37,99,235,0.15)"; }}
            onBlur={(e) => { e.target.style.borderColor = "#e5e7eb"; e.target.style.boxShadow = "0 1px 3px rgba(0,0,0,0.06)"; }}
          />
          <button
            onClick={handleSearch}
            style={{
              padding: "14px 28px",
              fontSize: 16,
              fontWeight: 600,
              background: "#2563eb",
              color: "#fff",
              border: "none",
              borderRadius: 10,
              cursor: "pointer",
              whiteSpace: "nowrap",
              boxShadow: "0 1px 3px rgba(37,99,235,0.3)",
              transition: "background 0.15s, box-shadow 0.15s",
            }}
            onMouseEnter={(e) => { e.currentTarget.style.background = "#1d4ed8"; e.currentTarget.style.boxShadow = "0 2px 8px rgba(37,99,235,0.4)"; }}
            onMouseLeave={(e) => { e.currentTarget.style.background = "#2563eb"; e.currentTarget.style.boxShadow = "0 1px 3px rgba(37,99,235,0.3)"; }}
          >
            {t("search.button")}
          </button>
        </div>

        {error && (
          <p style={{ color: "#dc2626", fontSize: 14, marginBottom: 16 }}>{error}</p>
        )}

        {/* Demo Cards */}
        <div style={{
          display: "flex",
          gap: 16,
          marginTop: 48,
          width: "100%",
          maxWidth: 720,
          flexWrap: "wrap",
          justifyContent: "center",
        }}>
          {DEMOS.map((demo) => (
            <button
              key={demo.repo}
              onClick={() => handleDemo(demo.repo)}
              style={{
                flex: "1 1 200px",
                maxWidth: 230,
                padding: "20px 16px",
                background: "#fff",
                border: "1px solid #e5e7eb",
                borderRadius: 14,
                cursor: "pointer",
                textAlign: "left",
                boxShadow: "0 1px 4px rgba(0,0,0,0.06)",
                transition: "box-shadow 0.2s, transform 0.2s, border-color 0.2s",
              }}
              onMouseEnter={(e) => { e.currentTarget.style.boxShadow = "0 6px 20px rgba(0,0,0,0.1)"; e.currentTarget.style.transform = "translateY(-2px)"; e.currentTarget.style.borderColor = "#93c5fd"; }}
              onMouseLeave={(e) => { e.currentTarget.style.boxShadow = "0 1px 4px rgba(0,0,0,0.06)"; e.currentTarget.style.transform = "translateY(0)"; e.currentTarget.style.borderColor = "#e5e7eb"; }}
            >
              <div style={{ fontSize: 28, marginBottom: 8 }}>{demo.icon}</div>
              <div style={{ fontSize: 15, fontWeight: 600, color: "#111827", marginBottom: 4 }}>
                {demo.title}
              </div>
              <div style={{ fontSize: 13, color: "#6b7280", lineHeight: 1.4 }}>
                {demo.description}
              </div>
            </button>
          ))}
        </div>
      </div>

      {/* Stats Footer */}
      <div style={{
        padding: "24px 20px",
        textAlign: "center",
        borderTop: "1px solid #e5e7eb",
        background: "#fff",
      }}>
        <div style={{
          display: "flex",
          justifyContent: "center",
          gap: 40,
          fontSize: 14,
          color: "#6b7280",
          flexWrap: "wrap",
        }}>
          <span><strong style={{ color: "#111827" }}>105K+</strong> {t("search.stats.repos")}</span>
          <span><strong style={{ color: "#111827" }}>125K+</strong> {t("search.stats.topics")}</span>
          <span><strong style={{ color: "#111827" }}>28</strong> {t("search.stats.ecosystems")}</span>
          <span><strong style={{ color: "#111827" }}>28K+</strong> {t("search.stats.connections")}</span>
        </div>
      </div>
    </div>
  );
}

import { useState } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { useI18n } from "../i18n";

export default function Navbar() {
  const navigate = useNavigate();
  const location = useLocation();
  const { t, lang, setLang } = useI18n();
  const [showContact, setShowContact] = useState(false);

  const isHome = location.pathname === "/";
  const isEcosystems = location.pathname.startsWith("/ecosystem");

  return (
    <nav style={{
      display: "flex",
      alignItems: "center",
      justifyContent: "space-between",
      padding: "12px 24px",
      background: "#fff",
      borderBottom: "1px solid #e5e7eb",
      position: "sticky",
      top: 0,
      zIndex: 100,
    }}>
      {/* Left: Logo + Nav Links */}
      <div style={{ display: "flex", alignItems: "center", gap: 24 }}>
        <button
          onClick={() => navigate("/")}
          style={{
            background: "none",
            border: "none",
            cursor: "pointer",
            fontSize: 18,
            fontWeight: 800,
            color: "#111827",
            padding: 0,
            letterSpacing: "-0.02em",
          }}
        >
          GitSense
        </button>

        <div style={{ display: "flex", gap: 4 }}>
          <button
            onClick={() => navigate("/")}
            style={{
              background: isHome ? "#eff6ff" : "none",
              border: "none",
              color: isHome ? "#2563eb" : "#6b7280",
              cursor: "pointer",
              fontSize: 14,
              fontWeight: isHome ? 600 : 400,
              padding: "6px 12px",
              borderRadius: 6,
              transition: "background 0.15s, color 0.15s",
            }}
            onMouseEnter={(e) => { if (!isHome) { e.currentTarget.style.background = "#f3f4f6"; e.currentTarget.style.color = "#111827"; } }}
            onMouseLeave={(e) => { if (!isHome) { e.currentTarget.style.background = "none"; e.currentTarget.style.color = "#6b7280"; } }}
          >
            {t("nav.home")}
          </button>

          <button
            onClick={() => navigate("/ecosystems")}
            style={{
              background: isEcosystems ? "#eff6ff" : "none",
              border: "none",
              color: isEcosystems ? "#2563eb" : "#6b7280",
              cursor: "pointer",
              fontSize: 14,
              fontWeight: isEcosystems ? 600 : 400,
              padding: "6px 12px",
              borderRadius: 6,
              transition: "background 0.15s, color 0.15s",
            }}
            onMouseEnter={(e) => { if (!isEcosystems) { e.currentTarget.style.background = "#f3f4f6"; e.currentTarget.style.color = "#111827"; } }}
            onMouseLeave={(e) => { if (!isEcosystems) { e.currentTarget.style.background = "none"; e.currentTarget.style.color = "#6b7280"; } }}
          >
            {t("nav.ecosystems")}
          </button>

          <button
            onClick={() => setShowContact(!showContact)}
            style={{
              background: "none",
              border: "none",
              color: "#6b7280",
              cursor: "pointer",
              fontSize: 14,
              fontWeight: 400,
              padding: "6px 12px",
              borderRadius: 6,
              transition: "background 0.15s, color 0.15s",
            }}
            onMouseEnter={(e) => { e.currentTarget.style.background = "#f3f4f6"; e.currentTarget.style.color = "#111827"; }}
            onMouseLeave={(e) => { e.currentTarget.style.background = "none"; e.currentTarget.style.color = "#6b7280"; }}
          >
            {t("nav.contact")}
          </button>
        </div>
      </div>

      {/* Right: Language Toggle */}
      <button
        onClick={() => setLang(lang === "en" ? "zh" : "en")}
        style={{
          background: "#f3f4f6",
          border: "1px solid #e5e7eb",
          borderRadius: 6,
          padding: "6px 12px",
          fontSize: 13,
          fontWeight: 500,
          color: "#374151",
          cursor: "pointer",
          transition: "background 0.15s",
          minWidth: 48,
          textAlign: "center",
        }}
        onMouseEnter={(e) => { e.currentTarget.style.background = "#e5e7eb"; }}
        onMouseLeave={(e) => { e.currentTarget.style.background = "#f3f4f6"; }}
      >
        {t("nav.lang")}
      </button>

      {/* Contact Popup */}
      {showContact && (
        <>
          <div onClick={() => setShowContact(false)} style={{ position: "fixed", inset: 0, zIndex: 199 }} />
          <div style={{
            position: "absolute",
            top: "100%",
            right: 24,
            marginTop: 8,
            background: "#fff",
            border: "1px solid #e5e7eb",
            borderRadius: 12,
            padding: "20px 24px",
            boxShadow: "0 8px 30px rgba(0,0,0,0.12)",
            zIndex: 200,
            minWidth: 260,
          }}>
            <div style={{ fontSize: 15, fontWeight: 600, color: "#111827", marginBottom: 16 }}>
              {t("nav.contact")}
            </div>

            <a
              href="https://github.com/stefall-code"
              target="_blank"
              rel="noopener noreferrer"
              style={{
                display: "flex",
                alignItems: "center",
                gap: 10,
                padding: "10px 12px",
                borderRadius: 8,
                textDecoration: "none",
                color: "#111827",
                background: "#f9fafb",
                marginBottom: 8,
                transition: "background 0.15s",
              }}
              onMouseEnter={(e) => { e.currentTarget.style.background = "#eff6ff"; }}
              onMouseLeave={(e) => { e.currentTarget.style.background = "#f9fafb"; }}
            >
              <svg width="20" height="20" viewBox="0 0 16 16" fill="#111827"><path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/></svg>
              <div>
                <div style={{ fontSize: 14, fontWeight: 600 }}>GitHub</div>
                <div style={{ fontSize: 12, color: "#6b7280" }}>stefall-code</div>
              </div>
            </a>

            <a
              href="mailto:2107327691@qq.com"
              style={{
                display: "flex",
                alignItems: "center",
                gap: 10,
                padding: "10px 12px",
                borderRadius: 8,
                textDecoration: "none",
                color: "#111827",
                background: "#f9fafb",
                transition: "background 0.15s",
              }}
              onMouseEnter={(e) => { e.currentTarget.style.background = "#eff6ff"; }}
              onMouseLeave={(e) => { e.currentTarget.style.background = "#f9fafb"; }}
            >
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#111827" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="2" y="4" width="20" height="16" rx="2"/><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/></svg>
              <div>
                <div style={{ fontSize: 14, fontWeight: 600 }}>{t("nav.email")}</div>
                <div style={{ fontSize: 12, color: "#6b7280" }}>2107327691@qq.com</div>
              </div>
            </a>
          </div>
        </>
      )}
    </nav>
  );
}

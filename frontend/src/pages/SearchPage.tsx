import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { parseRepoInput } from "../api";

export default function SearchPage() {
  const [input, setInput] = useState("");
  const [error, setError] = useState("");
  const navigate = useNavigate();

  const handleSearch = () => {
    setError("");
    const parsed = parseRepoInput(input);
    if (!parsed) {
      setError("Please enter a valid GitHub repo (owner/name or URL)");
      return;
    }
    navigate(`/result?owner=${encodeURIComponent(parsed.owner)}&repo=${encodeURIComponent(parsed.repo)}`);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") handleSearch();
  };

  return (
    <div style={{ maxWidth: 640, margin: "0 auto", padding: "80px 20px 0" }}>
      <h1 style={{ fontSize: 32, fontWeight: 700, marginBottom: 8, color: "#111827" }}>
        GitSense
      </h1>
      <p style={{ color: "#6b7280", marginBottom: 32, fontSize: 16 }}>
        Discover similar open-source repositories and technology ecosystems
      </p>

      <div style={{ display: "flex", gap: 8 }}>
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="e.g. facebook/react or https://github.com/facebook/react"
          style={{
            flex: 1,
            padding: "12px 16px",
            fontSize: 16,
            border: "1px solid #d1d5db",
            borderRadius: 8,
            outline: "none",
          }}
          onFocus={(e) => (e.target.style.borderColor = "#2563eb")}
          onBlur={(e) => (e.target.style.borderColor = "#d1d5db")}
        />
        <button
          onClick={handleSearch}
          style={{
            padding: "12px 24px",
            fontSize: 16,
            fontWeight: 600,
            background: "#2563eb",
            color: "#fff",
            border: "none",
            borderRadius: 8,
            cursor: "pointer",
          }}
        >
          Search
        </button>
      </div>

      {error && (
        <p style={{ color: "#dc2626", marginTop: 8, fontSize: 14 }}>{error}</p>
      )}

      <div style={{ marginTop: 48, color: "#9ca3af", fontSize: 13 }}>
        <p>Supported formats:</p>
        <ul style={{ marginTop: 4 }}>
          <li>owner/repo — e.g. openai/tiktoken</li>
          <li>GitHub URL — e.g. https://github.com/openai/tiktoken</li>
        </ul>
      </div>
    </div>
  );
}

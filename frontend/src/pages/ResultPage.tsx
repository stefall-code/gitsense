import { useEffect } from "react";
import { useSearchParams, useNavigate } from "react-router-dom";

export default function ResultPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const owner = searchParams.get("owner") || "";
  const repo = searchParams.get("repo") || "";

  useEffect(() => {
    if (owner && repo) {
      navigate(`/r/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}`, { replace: true });
    } else {
      navigate("/", { replace: true });
    }
  }, [owner, repo, navigate]);

  return (
    <div style={{ textAlign: "center", padding: 80, color: "#9ca3af" }}>
      Redirecting...
    </div>
  );
}

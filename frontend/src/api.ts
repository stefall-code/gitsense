import type { RecommendationResponse, TrendOverview } from "./types";

const API_BASE = import.meta.env.VITE_API_URL || "/api/v1";

export async function getRecommendations(
  owner: string,
  repo: string,
  options?: { limit?: number; strategy?: string; debug?: boolean }
): Promise<RecommendationResponse> {
  const params = new URLSearchParams();
  if (options?.limit) params.set("limit", String(options.limit));
  if (options?.strategy) params.set("strategy", options.strategy);
  if (options?.debug) params.set("debug", "true");

  const qs = params.toString() ? `?${params.toString()}` : "";
  const res = await fetch(
    `${API_BASE}/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}/recommendations${qs}`
  );

  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error?.message || `Request failed: ${res.status}`);
  }

  return res.json();
}

export async function getTrendOverview(
  window?: string
): Promise<TrendOverview> {
  const qs = window ? `?window=${window}` : "";
  const res = await fetch(`${API_BASE}/trends/overview${qs}`);

  if (!res.ok) {
    throw new Error(`Trend request failed: ${res.status}`);
  }

  return res.json();
}

export function parseRepoInput(input: string): { owner: string; repo: string } | null {
  const trimmed = input.trim();

  // URL format: https://github.com/owner/repo
  const urlMatch = trimmed.match(/github\.com\/([^/]+)\/([^/]+)/);
  if (urlMatch) {
    return { owner: urlMatch[1], repo: urlMatch[2] };
  }

  // owner/name format
  const parts = trimmed.split("/");
  if (parts.length === 2 && parts[0] && parts[1]) {
    return { owner: parts[0], repo: parts[1] };
  }

  return null;
}

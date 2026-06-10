// Minimal analytics for GitSense MVP
// Events: search, recommendation_click, share, ecosystem_open

interface AnalyticsEvent {
  event: string;
  properties: Record<string, string | number | boolean>;
  timestamp: string;
}

const SESSION_ID = Math.random().toString(36).substring(2, 10);
const QUEUE: AnalyticsEvent[] = [];
const FLUSH_INTERVAL = 5000;
const API_BASE = import.meta.env.VITE_API_URL || "/api/v1";

function flush() {
  if (QUEUE.length === 0) return;
  const events = QUEUE.splice(0, QUEUE.length);
  const payload = { session_id: SESSION_ID, events };

  // Fire-and-forget — never block UI
  fetch(`${API_BASE}/analytics/events`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  }).catch(() => {
    // Silently fail — analytics should never break the app
  });
}

// Auto-flush every 5 seconds
setInterval(flush, FLUSH_INTERVAL);

// Flush on page unload
if (typeof window !== "undefined") {
  window.addEventListener("beforeunload", flush);
}

export function track(event: string, properties: Record<string, string | number | boolean> = {}) {
  QUEUE.push({
    event,
    properties: { ...properties, session_id: SESSION_ID },
    timestamp: new Date().toISOString(),
  });

  // Flush immediately if queue is getting large
  if (QUEUE.length >= 10) {
    flush();
  }
}

// Convenience helpers
export function trackSearch(query: string) {
  track("search", { query });
}

export function trackRecommendationClick(repo: string, fromRepo: string, rank: number) {
  track("recommendation_click", { repo, from_repo: fromRepo, rank });
}

export function trackShare(repo: string) {
  track("share", { repo });
}

export function trackEcosystemOpen(ecosystem: string, fromRepo: string) {
  track("ecosystem_open", { ecosystem, from_repo: fromRepo });
}

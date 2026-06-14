import { createContext, useContext, useState, type ReactNode } from "react";

type Lang = "en" | "zh";

interface I18nValue {
  lang: Lang;
  setLang: (l: Lang) => void;
  t: (key: string) => string;
  tr: (reason: string) => string;
}

const translations: Record<Lang, Record<string, string>> = {
  en: {
    // Nav
    "nav.home": "Home",
    "nav.ecosystems": "Ecosystems",
    "nav.lang": "中文",
    "nav.contact": "Contact",
    "nav.email": "Email",

    // Search Page
    "search.title": "GitSense",
    "search.subtitle1": "GitHub Search tells you what exists.",
    "search.subtitle2": "GitSense tells you what to use.",
    "search.placeholder": "Enter a GitHub repo — e.g. facebook/react",
    "search.button": "Discover",
    "search.error": "Please enter a valid GitHub repo (owner/name or URL)",
    "search.demo.ai.title": "AI Agent Ecosystem",
    "search.demo.ai.desc": "Explore the AI agent landscape: LangChain, LlamaIndex, CrewAI, AutoGen",
    "search.demo.go.title": "Go Web Frameworks",
    "search.demo.go.desc": "Compare Go web frameworks: Gin, Echo, Fiber, Chi",
    "search.demo.data.title": "Python Data Science",
    "search.demo.data.desc": "Navigate the Python data science stack: NumPy, SciPy, Scikit-learn",
    "search.stats.repos": "Repositories",
    "search.stats.topics": "Technology Topics",
    "search.stats.ecosystems": "Curated Ecosystems",
    "search.stats.connections": "Repository Connections",

    // Repo Profile
    "repo.loading": "Discovering ecosystem...",
    "repo.notFound.title": "Repository not found",
    "repo.notFound.desc": "on GitHub. Check the spelling and try again.",
    "repo.notFound.button": "Try another repo",
    "repo.notIndexed.title": "Not yet indexed",
    "repo.notIndexed.desc": "exists but hasn't been indexed yet. We're constantly adding more repos.",
    "repo.notIndexed.explore": "Explore other repos",
    "repo.notIndexed.github": "View on GitHub",
    "repo.error.title": "Something went wrong",
    "repo.error.desc": "We encountered an error while discovering this repo. Please try again.",
    "repo.error.retry": "Retry",
    "repo.newSearch": "New Search",
    "repo.share": "Share",
    "repo.share.copied": "Link copied!",
    "repo.share.failed": "Failed to copy",
    "repo.noEcosystem": "Ecosystem not yet classified for this repo",
    "repo.similar": "similar repositories",
    "repo.noSimilar": "No similar repositories found yet",
    "repo.noSimilar.hint": "Try exploring the ecosystem above, or search for another repo",

    // Ecosystem List
    "ecoList.title": "All Ecosystems",
    "ecoList.back": "Back to Search",
    "ecoList.loading": "Loading ecosystems...",
    "ecoList.repos": "repos",
    "ecoList.categories": "categories",

    // Ecosystem Detail
    "ecoDetail.back": "All Ecosystems",
    "ecoDetail.loading": "Loading ecosystem...",
    "ecoDetail.trendScore": "Trend Score",
    "ecoDetail.growthRate": "Growth Rate",
    "ecoDetail.categories": "Categories",
    "ecoDetail.topRepos": "Top Repositories",

    // Tech Stack
    "stack.top": "Top",
    "stack.trending": "Trending",

    // Trend
    "trend.rising": "rising",
    "trend.declining": "declining",
    "trend.stable": "stable",
    "trend.title": "Trending in",

    // Ecosystem Header
    "ecoHeader.repos": "repos",
    "ecoHeader.subcategory": "Subcategory",
    "ecoHeader.viewAll": "View ecosystem →",
  },
  zh: {
    // Nav
    "nav.home": "首页",
    "nav.ecosystems": "生态",
    "nav.lang": "EN",
    "nav.contact": "联系开发者",
    "nav.email": "邮箱",

    // Search Page
    "search.title": "GitSense",
    "search.subtitle1": "GitHub 搜索告诉你有什么。",
    "search.subtitle2": "GitSense 告诉你该用什么。",
    "search.placeholder": "输入 GitHub 仓库 — 例如 facebook/react",
    "search.button": "发现",
    "search.error": "请输入有效的 GitHub 仓库（owner/name 或 URL）",
    "search.demo.ai.title": "AI Agent 生态",
    "search.demo.ai.desc": "探索 AI Agent 版图：LangChain、LlamaIndex、CrewAI、AutoGen",
    "search.demo.go.title": "Go Web 框架",
    "search.demo.go.desc": "对比 Go Web 框架：Gin、Echo、Fiber、Chi",
    "search.demo.data.title": "Python 数据科学",
    "search.demo.data.desc": "导航 Python 数据科学栈：NumPy、SciPy、Scikit-learn",
    "search.stats.repos": "开源仓库",
    "search.stats.topics": "技术标签",
    "search.stats.ecosystems": "精选生态",
    "search.stats.connections": "仓库连接",

    // Repo Profile
    "repo.loading": "正在发现生态...",
    "repo.notFound.title": "未找到仓库",
    "repo.notFound.desc": "在 GitHub 上不存在。请检查拼写后重试。",
    "repo.notFound.button": "搜索其他仓库",
    "repo.notIndexed.title": "尚未收录",
    "repo.notIndexed.desc": "已存在但尚未被收录。我们正在持续添加更多仓库。",
    "repo.notIndexed.explore": "探索其他仓库",
    "repo.notIndexed.github": "在 GitHub 查看",
    "repo.error.title": "出了点问题",
    "repo.error.desc": "在发现该仓库生态时遇到了错误，请重试。",
    "repo.error.retry": "重试",
    "repo.newSearch": "新搜索",
    "repo.share": "分享",
    "repo.share.copied": "链接已复制！",
    "repo.share.failed": "复制失败",
    "repo.noEcosystem": "该仓库尚未分类到任何生态",
    "repo.similar": "个相似仓库",
    "repo.noSimilar": "暂未找到相似仓库",
    "repo.noSimilar.hint": "试试浏览上方的生态，或搜索其他仓库",

    // Ecosystem List
    "ecoList.title": "全部生态",
    "ecoList.back": "返回搜索",
    "ecoList.loading": "加载生态列表...",
    "ecoList.repos": "个仓库",
    "ecoList.categories": "个分类",

    // Ecosystem Detail
    "ecoDetail.back": "全部生态",
    "ecoDetail.loading": "加载生态详情...",
    "ecoDetail.trendScore": "趋势分数",
    "ecoDetail.growthRate": "增长率",
    "ecoDetail.categories": "分类",
    "ecoDetail.topRepos": "热门仓库",

    // Tech Stack
    "stack.top": "热门",
    "stack.trending": "趋势",

    // Trend
    "trend.rising": "上升",
    "trend.declining": "下降",
    "trend.stable": "稳定",
    "trend.title": "趋势",

    // Ecosystem Header
    "ecoHeader.repos": "个仓库",
    "ecoHeader.subcategory": "子分类",
    "ecoHeader.viewAll": "查看生态 →",

    // Reasons (pattern-based translation)
    "reason.highSemantic": "高语义相似度 (embedding_score=${score})",
    "reason.moderateSemantic": "中等语义相似度 (embedding_score=${score})",
    "reason.strongGraph": "强图连接 (graph_score=${score})",
    "reason.directGraph": "直接图相似连接",
    "reason.twoHopGraph": "通过2跳图路径连接",
    "reason.risingTrend": "生态内上升趋势 (trend_score=${score})",
    "reason.decliningTrend": "下降趋势信号 (trend_score=${score})",
    "reason.highPopularity": "高人气 (popularity_score=${score}, stars=${stars})",
    "reason.popularProject": "热门项目 (stars=${stars})",
    "reason.strongTopic": "强 Topic 重叠 (topic_score=${score})",
    "reason.sharedTopics": "共享 Topics: ${topics} (topic_score=${score})",
    "reason.sameEcosystem": "同一技术生态",
    "reason.similarityScore": "相似度分数: ${score}",
  },
};

// Translate a backend-generated reason string to current language
const reasonPatterns: { pattern: RegExp; key: string }[] = [
  { pattern: /^High semantic similarity \(embedding_score=([\d.]+)\)$/, key: "reason.highSemantic" },
  { pattern: /^Moderate semantic similarity \(embedding_score=([\d.]+)\)$/, key: "reason.moderateSemantic" },
  { pattern: /^Strong graph connectivity \(graph_score=([\d.]+)\)$/, key: "reason.strongGraph" },
  { pattern: /^Direct graph similarity connection$/, key: "reason.directGraph" },
  { pattern: /^Connected via 2-hop graph path$/, key: "reason.twoHopGraph" },
  { pattern: /^Rising trend in ecosystem \(trend_score=([\d.]+)\)$/, key: "reason.risingTrend" },
  { pattern: /^Declining trend signal \(trend_score=([\d.]+)\)$/, key: "reason.decliningTrend" },
  { pattern: /^High popularity \(popularity_score=([\d.]+), stars=(\d+)\)$/, key: "reason.highPopularity" },
  { pattern: /^Popular project \(stars=(\d+)\)$/, key: "reason.popularProject" },
  { pattern: /^Strong topic overlap \(topic_score=([\d.]+)\)$/, key: "reason.strongTopic" },
  { pattern: /^Shared topics: \[([^\]]+)\] \(topic_score=([\d.]+)\)$/, key: "reason.sharedTopics" },
  { pattern: /^Same technology ecosystem$/, key: "reason.sameEcosystem" },
  { pattern: /^Similarity score: ([\d.]+)$/, key: "reason.similarityScore" },
];

const I18nContext = createContext<I18nValue | null>(null);

export function I18nProvider({ children }: { children: ReactNode }) {
  const [lang, setLang] = useState<Lang>(() => {
    const saved = localStorage.getItem("gitsense-lang");
    return (saved === "zh" || saved === "en") ? saved : "en";
  });

  const handleSetLang = (l: Lang) => {
    setLang(l);
    localStorage.setItem("gitsense-lang", l);
  };

  const t = (key: string): string => {
    return translations[lang][key] || translations.en[key] || key;
  };

  const tr = (reason: string): string => {
    if (lang === "en") return reason;
    for (const { pattern, key } of reasonPatterns) {
      const match = reason.match(pattern);
      if (match) {
        let result = translations.zh[key] || reason;
        if (match[1] !== undefined) result = result.replace("${score}", match[1]).replace("${stars}", match[2] || "").replace("${topics}", match[1]);
        if (match[2] !== undefined) result = result.replace("${stars}", match[2]).replace("${topics}", match[1]);
        return result;
      }
    }
    return reason;
  };

  return (
    <I18nContext.Provider value={{ lang, setLang: handleSetLang, t, tr }}>
      {children}
    </I18nContext.Provider>
  );
}

export function useI18n() {
  const ctx = useContext(I18nContext);
  if (!ctx) throw new Error("useI18n must be used within I18nProvider");
  return ctx;
}

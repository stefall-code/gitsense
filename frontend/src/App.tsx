import { BrowserRouter, Routes, Route } from "react-router-dom";
import { I18nProvider } from "./i18n";
import Navbar from "./components/Navbar";
import SearchPage from "./pages/SearchPage";
import ResultPage from "./pages/ResultPage";
import EcosystemsListPage from "./pages/EcosystemsListPage";
import EcosystemDetailPage from "./pages/EcosystemDetailPage";
import RepoProfilePage from "./pages/RepoProfilePage";

export default function App() {
  return (
    <I18nProvider>
      <BrowserRouter>
        <div style={{ minHeight: "100vh", background: "#f9fafb" }}>
          <Navbar />
          <Routes>
            <Route path="/" element={<SearchPage />} />
            <Route path="/result" element={<ResultPage />} />
            <Route path="/r/:owner/:repo" element={<RepoProfilePage />} />
            <Route path="/ecosystems" element={<EcosystemsListPage />} />
            <Route path="/ecosystem/:name" element={<EcosystemDetailPage />} />
          </Routes>
          <footer style={{
            padding: "16px 20px",
            textAlign: "center",
            fontSize: 12,
            color: "#9ca3af",
            background: "#f9fafb",
          }}>
            <a
              href="https://beian.miit.gov.cn/"
              target="_blank"
              rel="noopener noreferrer"
              style={{ color: "#9ca3af", textDecoration: "none" }}
            >
              粤ICP备2026081993号
            </a>
          </footer>
        </div>
      </BrowserRouter>
    </I18nProvider>
  );
}

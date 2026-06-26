import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { AdminLayout } from "./layouts/AdminLayout";
import { PublicLayout } from "./layouts/PublicLayout";
import { UserLayout } from "./layouts/UserLayout";
import { MockStateProvider } from "./mock/MockState";
import { AdminOverviewPage } from "./routes/AdminOverviewPage";
import { AdminProgramsPage } from "./routes/AdminProgramsPage";
import { CollectionEditorPage } from "./routes/CollectionEditorPage";
import { CollectionsPage } from "./routes/CollectionsPage";
import { ComponentShowcasePage } from "./routes/ComponentShowcasePage";
import { ForgotPasswordPage } from "./routes/ForgotPasswordPage";
import { HomePage } from "./routes/HomePage";
import { LoginPage } from "./routes/LoginPage";
import { ProgramDetailPage } from "./routes/ProgramDetailPage";
import { ProgramsPage } from "./routes/ProgramsPage";
import { RegisterPage } from "./routes/RegisterPage";
import { RegisterVerifyPage } from "./routes/RegisterVerifyPage";
import { SubscribePage } from "./routes/SubscribePage";
import "./styles/tokens.css";
import "./styles/global.css";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <MockStateProvider>
      <BrowserRouter>
        <Routes>
          <Route element={<PublicLayout />}>
            <Route path="/" element={<HomePage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/register/verify" element={<RegisterVerifyPage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/forgot-password" element={<ForgotPasswordPage />} />
            <Route path="/components" element={<ComponentShowcasePage />} />
          </Route>
          <Route element={<UserLayout />}>
            <Route path="/programs" element={<ProgramsPage />} />
            <Route path="/programs/:programId" element={<ProgramDetailPage />} />
            <Route path="/collections" element={<CollectionsPage />} />
            <Route path="/collections/:collectionId" element={<CollectionEditorPage />} />
            <Route path="/collections/:collectionId/subscribe" element={<SubscribePage />} />
          </Route>
          <Route path="/admin" element={<AdminLayout />}>
            <Route index element={<AdminOverviewPage />} />
            <Route path="programs" element={<AdminProgramsPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </MockStateProvider>
  </StrictMode>
);

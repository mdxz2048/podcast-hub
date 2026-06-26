import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { AdminRouteGuard } from "./auth/AdminRouteGuard";
import { AuthProvider } from "./auth/AuthProvider";
import { AdminLayout } from "./layouts/AdminLayout";
import { PublicLayout } from "./layouts/PublicLayout";
import { UserLayout } from "./layouts/UserLayout";
import { MockStateProvider } from "./mock/MockState";
import { AdminOverviewPage } from "./routes/AdminOverviewPage";
import { AdminConnectorDetailPage } from "./routes/AdminConnectorDetailPage";
import { AdminConnectorNewPage } from "./routes/AdminConnectorNewPage";
import { AdminConnectorsPage } from "./routes/AdminConnectorsPage";
import { AdminImportJobDetailPage } from "./routes/AdminImportJobDetailPage";
import { AdminImportJobsPage } from "./routes/AdminImportJobsPage";
import { AdminProgramDetailPage } from "./routes/AdminProgramDetailPage";
import { AdminProgramsPage } from "./routes/AdminProgramsPage";
import { AdminReviewsPage } from "./routes/AdminReviewsPage";
import { AdminSecretsPage } from "./routes/AdminSecretsPage";
import { AdminSourceDetailPage } from "./routes/AdminSourceDetailPage";
import { AdminSourceNewPage } from "./routes/AdminSourceNewPage";
import { AdminSourcesPage } from "./routes/AdminSourcesPage";
import { AdminUsersPage } from "./routes/AdminUsersPage";
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
import { ResetPasswordPage } from "./routes/ResetPasswordPage";
import { SubscribePage } from "./routes/SubscribePage";
import "./styles/tokens.css";
import "./styles/global.css";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <MockStateProvider>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route element={<PublicLayout />}>
              <Route path="/" element={<HomePage />} />
              <Route path="/register" element={<RegisterPage />} />
              <Route path="/register/verify" element={<RegisterVerifyPage />} />
              <Route path="/login" element={<LoginPage />} />
              <Route path="/forgot-password" element={<ForgotPasswordPage />} />
              <Route path="/reset-password" element={<ResetPasswordPage />} />
              <Route path="/components" element={<ComponentShowcasePage />} />
            </Route>
            <Route element={<UserLayout />}>
              <Route path="/programs" element={<ProgramsPage />} />
              <Route path="/programs/:programId" element={<ProgramDetailPage />} />
              <Route path="/collections" element={<CollectionsPage />} />
              <Route path="/collections/:collectionId" element={<CollectionEditorPage />} />
              <Route path="/collections/:collectionId/subscribe" element={<SubscribePage />} />
            </Route>
            <Route path="/admin" element={<AdminRouteGuard><AdminLayout /></AdminRouteGuard>}>
              <Route index element={<AdminOverviewPage />} />
              <Route path="programs" element={<AdminProgramsPage />} />
              <Route path="programs/:programId" element={<AdminProgramDetailPage />} />
              <Route path="connectors" element={<AdminConnectorsPage />} />
              <Route path="connectors/new" element={<AdminConnectorNewPage />} />
              <Route path="connectors/:connectorId" element={<AdminConnectorDetailPage />} />
              <Route path="sources" element={<AdminSourcesPage />} />
              <Route path="sources/new" element={<AdminSourceNewPage />} />
              <Route path="sources/:sourceId" element={<AdminSourceDetailPage />} />
              <Route path="secrets" element={<AdminSecretsPage />} />
              <Route path="import-jobs" element={<AdminImportJobsPage />} />
              <Route path="import-jobs/:jobId" element={<AdminImportJobDetailPage />} />
              <Route path="reviews" element={<AdminReviewsPage />} />
              <Route path="users" element={<AdminUsersPage />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </MockStateProvider>
  </StrictMode>
);

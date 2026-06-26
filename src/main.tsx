import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { AdminLayout } from "./layouts/AdminLayout";
import { PublicLayout } from "./layouts/PublicLayout";
import { UserLayout } from "./layouts/UserLayout";
import { AdminOverviewPage } from "./routes/AdminOverviewPage";
import { AdminProgramsPage } from "./routes/AdminProgramsPage";
import { ComponentShowcasePage } from "./routes/ComponentShowcasePage";
import { HomePage } from "./routes/HomePage";
import { LoginPage } from "./routes/LoginPage";
import { ProgramsPage } from "./routes/ProgramsPage";
import { RegisterPage } from "./routes/RegisterPage";
import "./styles/tokens.css";
import "./styles/global.css";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route element={<PublicLayout />}>
          <Route path="/" element={<HomePage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/components" element={<ComponentShowcasePage />} />
        </Route>
        <Route element={<UserLayout />}>
          <Route path="/programs" element={<ProgramsPage />} />
        </Route>
        <Route path="/admin" element={<AdminLayout />}>
          <Route index element={<AdminOverviewPage />} />
          <Route path="programs" element={<AdminProgramsPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  </StrictMode>
);


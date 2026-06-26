import type { ReactNode } from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "./AuthProvider";
import { LoadingState, PermissionDeniedState } from "../components/StateBlocks";

export function AdminRouteGuard({ children }: { children: ReactNode }) {
  const { user, isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return <LoadingState title="正在恢复登录状态" />;
  }
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  if (user?.role !== "admin") {
    return (
      <section className="mx-auto max-w-4xl px-5 py-10">
        <PermissionDeniedState />
      </section>
    );
  }
  return <>{children}</>;
}

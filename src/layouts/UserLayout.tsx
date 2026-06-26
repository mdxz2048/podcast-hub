import { Folder, Library, Radio, UserRound } from "lucide-react";
import { Link, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../auth/AuthProvider";
import { Button } from "../components/Button";

export function UserLayout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-canvas">
      <header className="border-b border-border bg-surface">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-5 py-4">
          <Link to="/" className="flex items-center gap-3 font-semibold">
            <span className="grid h-9 w-9 place-items-center rounded-md bg-action text-white">
              <Library className="h-5 w-5" />
            </span>
            Podcast Hub
          </Link>
          <nav className="flex items-center gap-2 text-sm">
            <Link className="rounded-md bg-subtle px-3 py-2 font-medium" to="/programs">
              节目
            </Link>
            <Link className="grid h-9 w-9 place-items-center rounded-md border border-border text-secondary sm:hidden" to="/collections" aria-label="合集">
              <Folder className="h-4 w-4" />
            </Link>
            <Link className="hidden items-center gap-2 rounded-md px-3 py-2 text-secondary hover:bg-subtle sm:flex" to="/collections">
              <Folder className="h-4 w-4" /> 合集
            </Link>
            <span className="hidden items-center gap-2 rounded-md px-3 py-2 text-secondary sm:flex">
              <Radio className="h-4 w-4" /> RSS
            </span>
            <span className="hidden items-center gap-2 rounded-md border border-border px-3 py-2 text-secondary md:inline-flex">
              <UserRound className="h-4 w-4" />
              <span className="max-w-40 truncate">{user?.email ?? "访客"}</span>
            </span>
            <Button
              type="button"
              variant="ghost"
              className="min-h-9 px-3"
              onClick={() => {
                void logout().then(() => navigate("/login"));
              }}
            >
              退出
            </Button>
          </nav>
        </div>
      </header>
      <main className="mx-auto max-w-7xl px-5 py-8">
        <Outlet />
      </main>
    </div>
  );
}

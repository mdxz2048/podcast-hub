import { Library, LogIn } from "lucide-react";
import { Link, Outlet, useLocation } from "react-router-dom";
import { Button } from "../components/Button";

export function PublicLayout() {
  const location = useLocation();
  const isAuth = location.pathname === "/login" || location.pathname === "/register";

  return (
    <div className="min-h-screen">
      <header className="mx-auto flex max-w-7xl items-center justify-between px-5 py-5">
        <Link to="/" className="flex items-center gap-3 font-semibold">
          <span className="grid h-9 w-9 place-items-center rounded-md bg-action text-white">
            <Library className="h-5 w-5" />
          </span>
          Podcast Hub
        </Link>
        <nav className="flex items-center gap-2">
          <Link className="hidden rounded-md px-3 py-2 text-sm text-secondary hover:bg-subtle sm:inline-flex" to="/programs">
            节目
          </Link>
          {!isAuth ? (
            <Link to="/login">
              <Button variant="secondary" icon={<LogIn className="h-4 w-4" />}>登录</Button>
            </Link>
          ) : null}
        </nav>
      </header>
      <main>
        <Outlet />
      </main>
    </div>
  );
}

import { Activity, ClipboardList, Inbox, KeyRound, Library, RadioTower, Settings, ShieldCheck, UsersRound } from "lucide-react";
import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../auth/AuthProvider";
import { Button } from "../components/Button";

const nav = [
  { to: "/admin", label: "概览", icon: Activity },
  { to: "/admin/programs", label: "节目", icon: Library },
  { to: "/admin/connectors", label: "连接器", icon: RadioTower },
  { to: "/admin/sources", label: "来源", icon: ClipboardList },
  { to: "/admin/secrets", label: "Secret", icon: KeyRound },
  { to: "/admin/import-jobs", label: "导入任务", icon: ClipboardList },
  { to: "/admin/staging", label: "待审核区", icon: Inbox },
  { to: "/admin/reviews", label: "审核", icon: ShieldCheck },
  { to: "/admin/users", label: "用户", icon: UsersRound },
  { to: "/admin/settings", label: "设置", icon: Settings }
];

export function AdminLayout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-canvas lg:grid lg:grid-cols-[248px_1fr]">
      <aside className="border-b border-border bg-surface lg:min-h-screen lg:border-b-0 lg:border-r">
        <div className="flex items-center justify-between px-5 py-4 lg:block">
          <div className="flex items-center gap-3 font-semibold">
            <span className="grid h-9 w-9 place-items-center rounded-md bg-action text-white">
              <ShieldCheck className="h-5 w-5" />
            </span>
            管理后台
          </div>
          <p className="hidden pt-2 text-xs text-muted lg:block">管理员权限链路验证页</p>
        </div>
        <div className="px-5 pb-3 text-xs text-secondary lg:pt-1">
          <div className="truncate">当前管理员：{user?.email ?? "未识别"}</div>
          <Button
            type="button"
            variant="ghost"
            className="mt-2 h-8 px-2 text-xs"
            onClick={() => {
              void logout().then(() => navigate("/login"));
            }}
          >
            退出登录
          </Button>
        </div>
        <nav className="flex gap-1 overflow-x-auto px-3 pb-3 lg:grid lg:px-3 lg:pb-0">
          {nav.map((item) => {
            const Icon = item.icon;
            return (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.to === "/admin"}
                className={({ isActive }) =>
                  `flex min-w-fit items-center gap-2 rounded-md px-3 py-2 text-sm ${
                    isActive ? "bg-subtle font-semibold text-primary" : "text-secondary hover:bg-subtle"
                  }`
                }
              >
                <Icon className="h-4 w-4" />
                {item.label}
              </NavLink>
            );
          })}
        </nav>
      </aside>
      <main className="px-5 py-6 lg:px-8">
        <Outlet />
      </main>
    </div>
  );
}

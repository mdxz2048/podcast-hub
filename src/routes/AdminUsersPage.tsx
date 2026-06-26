import { Eye, Radio, ShieldOff, ShieldCheck } from "lucide-react";
import { useState } from "react";
import { Link } from "react-router-dom";
import { useSearchParams } from "react-router-dom";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Dialog } from "../components/Dialog";
import { Drawer } from "../components/Drawer";
import { SearchBar, Select } from "../components/Form";
import { PageHeader } from "../components/PageShell";
import { EmptyState, PermissionDeniedState, SuccessFeedback } from "../components/StateBlocks";
import { adminUsers } from "../mock/data";
import { useMockState } from "../mock/MockState";
import { responsibilityLabel, rssTokenStateLabel, userRoleLabel, userStatusLabel } from "../utils/labels";

export function AdminUsersPage() {
  const [params] = useSearchParams();
  const { showToast } = useMockState();
  const [pendingAction, setPendingAction] = useState<{ type: "revoke_rss" | "toggle_suspend"; userId: string } | null>(null);
  const state = params.get("state");
  if (state === "denied") return <PermissionDeniedState />;

  const longTextUsers = adminUsers.map((user, index) => (
    index === 2
      ? {
        ...user,
        email: "listener.with.an.extremely.long.address.for.visual.acceptance.and.mobile.overflow.check@example.invalid",
        displayName: "普通用户（超长邮箱样例）",
        accessSummary: "用于验证用户列表在超长邮箱和超长说明场景下不会出现布局错位或横向溢出。"
      }
      : user
  ));
  const visible = state === "empty" ? [] : state === "long" ? longTextUsers : adminUsers;
  const activeUser = visible.find((user) => user.id === params.get("drawer"));
  const actionTarget = pendingAction ? adminUsers.find((user) => user.id === pendingAction.userId) : undefined;

  function action(title: string) {
    showToast({ tone: "success", title, message: "用户与访问权限操作已写入模拟状态，未调用真实接口。" });
  }

  return (
    <div className="grid gap-6">
      <PageHeader eyebrow="用户与访问权限" title="查看账号状态、角色、职责标签和 RSS 访问风险" />
      {params.get("toast") === "success" ? <SuccessFeedback message="模拟权限操作已完成。" /> : null}
      <div className="grid gap-3 rounded-lg border border-border bg-surface p-4 md:grid-cols-[1fr_220px]">
        <SearchBar placeholder="搜索邮箱或显示名" />
        <Select label="用户状态" options={["全部状态", "活跃", "待邮箱验证", "已暂停", "已删除"]} />
      </div>
      {visible.length === 0 ? <EmptyState title="没有用户记录" /> : (
        <div className="grid gap-4">
          {visible.map((user) => (
            <article key={user.id} className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
              <div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
                <div className="min-w-0">
                  <div className="flex flex-wrap gap-2">
                    <Badge tone={user.status === "active" ? "success" : user.status === "suspended" ? "danger" : "warning"}>{userStatusLabel[user.status]}</Badge>
                    <Badge>{userRoleLabel[user.role]}</Badge>
                    {user.responsibilityLabels.map((label) => <Badge key={label} tone="info">{responsibilityLabel[label]}</Badge>)}
                  </div>
                  <h2 className="mt-3 text-lg font-semibold">{user.displayName}</h2>
                  <p className="mt-1 break-all font-mono text-sm text-secondary">{user.email}</p>
                  <p className="mt-2 text-sm text-secondary">{user.accessSummary}</p>
                </div>
                <div className="grid gap-2 sm:grid-cols-2 xl:min-w-[360px]">
                  <Info label="可访问节目" value={`${user.accessibleProgramCount} 个`} />
                  <Info label="私有 RSS" value={user.privateRssState === "active" ? "活跃" : user.privateRssState === "revoked" ? "已撤销" : "暂停"} />
                  <Info label="最近活动" value={user.lastActivity} />
                  <Info label="职责说明" value={user.responsibilityLabels.length > 0 ? "职责标签不是账号角色" : "无管理职责"} />
                </div>
              </div>
              <div className="mt-4 flex flex-wrap gap-2">
                <Link to={`/admin/users?drawer=${user.id}`}>
                  <Button variant="secondary" icon={<Eye className="h-4 w-4" />}>查看访问权限</Button>
                </Link>
                <Button variant="secondary" icon={<Radio className="h-4 w-4" />} onClick={() => setPendingAction({ type: "revoke_rss", userId: user.id })}>撤销 RSS Token</Button>
                <Button variant={user.status === "suspended" ? "secondary" : "danger"} icon={user.status === "suspended" ? <ShieldCheck className="h-4 w-4" /> : <ShieldOff className="h-4 w-4" />} onClick={() => setPendingAction({ type: "toggle_suspend", userId: user.id })}>
                  {user.status === "suspended" ? "恢复用户" : "暂停用户"}
                </Button>
              </div>
            </article>
          ))}
        </div>
      )}
      <Drawer open={Boolean(activeUser)} title="访问权限详情" onClose={() => window.history.back()}>
        {activeUser ? (
          <div className="grid gap-4">
            <div>
              <p className="text-xs text-muted">账号角色</p>
              <p className="font-semibold">{userRoleLabel[activeUser.role]}</p>
              <p className="mt-2 text-sm text-secondary">System Owner、Operator、Reviewer 是职责标签，不是独立账号角色。</p>
            </div>
            <div className="rounded-lg border border-border bg-subtle p-4">
              <p>可访问节目：{activeUser.accessibleProgramCount} 个</p>
              <p>私有 RSS：{rssTokenStateLabel[activeUser.privateRssState]}</p>
              <p>状态：{userStatusLabel[activeUser.status]}</p>
            </div>
            <div className="rounded-lg border border-warning/30 bg-warning/5 p-4 text-sm text-secondary">
              撤销节目权限或暂停用户后，真实 RSS 请求必须实时失效。M0.2B 只展示静态状态。
            </div>
          </div>
        ) : null}
      </Drawer>
      <Dialog
        open={Boolean(actionTarget && pendingAction)}
        title={pendingAction?.type === "revoke_rss" ? "撤销 RSS Token" : actionTarget?.status === "suspended" ? "恢复用户" : "暂停用户"}
        description={pendingAction?.type === "revoke_rss"
          ? `确认撤销「${actionTarget?.displayName ?? ""}」的私有 RSS Token？`
          : actionTarget?.status === "suspended"
            ? `确认恢复「${actionTarget?.displayName ?? ""}」账号？`
            : `确认暂停「${actionTarget?.displayName ?? ""}」账号？暂停后私有 RSS 应立即失效。`}
        confirmLabel={pendingAction?.type === "revoke_rss" ? "确认撤销" : actionTarget?.status === "suspended" ? "确认恢复" : "确认暂停"}
        onCancel={() => setPendingAction(null)}
        onConfirm={() => {
          if (!actionTarget || !pendingAction) return;
          action(pendingAction.type === "revoke_rss" ? "RSS Token 已撤销" : actionTarget.status === "suspended" ? "用户已恢复" : "用户已暂停");
          setPendingAction(null);
        }}
      />
    </div>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border border-border bg-subtle p-3">
      <p className="text-xs text-muted">{label}</p>
      <p className="mt-1 text-sm font-medium">{value}</p>
    </div>
  );
}

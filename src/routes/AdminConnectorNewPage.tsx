import { CheckCircle2, FileArchive, ShieldAlert } from "lucide-react";
import { useState } from "react";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Input } from "../components/Form";
import { PageHeader } from "../components/PageShell";
import { SuccessFeedback } from "../components/StateBlocks";
import { useMockState } from "../mock/MockState";

const steps = ["选择接入方式", "模拟包信息", "Manifest 摘要", "权限审核"];

export function AdminConnectorNewPage() {
  const [step, setStep] = useState(0);
  const [registered, setRegistered] = useState(false);
  const { showToast } = useMockState();

  function finish() {
    setRegistered(true);
    showToast({ tone: "success", title: "已登记", message: "模拟 Connector 已登记，未读取或上传真实 ZIP。" });
  }

  return (
    <div className="grid gap-6">
      <PageHeader eyebrow="Connector 注册向导" title="静态验证未来上传 Connector 包的体验" />
      <section className="rounded-lg border border-warning/30 bg-warning/5 p-4 text-sm text-secondary">
        当前为静态演示：不读取、解压、执行或上传真实 ZIP；不接受真实凭据；不申请真实网络权限。
      </section>
      {registered ? <SuccessFeedback message="模拟登记完成。真实版本仍需要包校验、人工审核和隔离执行策略。" /> : null}
      <div className="grid gap-3 md:grid-cols-4">
        {steps.map((label, index) => (
          <button key={label} className={`rounded-lg border p-4 text-left ${step === index ? "border-action bg-subtle" : "border-border bg-surface"}`} onClick={() => setStep(index)}>
            <p className="text-xs text-muted">步骤 {index + 1}</p>
            <p className="font-semibold">{label}</p>
          </button>
        ))}
      </div>
      <section className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
        {step === 0 ? (
          <div className="grid gap-4 md:grid-cols-3">
            {["平台内建 Native RSS Importer", "Python Connector ZIP", "人工导入工作流"].map((item) => (
              <div key={item} className="rounded-lg border border-border p-4">
                <FileArchive className="mb-3 h-5 w-5 text-action" />
                <h2 className="font-semibold">{item}</h2>
                <p className="mt-2 text-sm text-secondary">仅用于选择接入方式，不上传真实文件。</p>
              </div>
            ))}
          </div>
        ) : null}
        {step === 1 ? (
          <div className="grid gap-4">
            <Input label="模拟包名称" defaultValue="partner-archive-connector-1.5.0.zip" />
            <div className="rounded-lg border border-dashed border-strong bg-subtle p-5 text-sm text-secondary">
              这里展示模拟文件名，不打开本地文件选择器，不读取磁盘，不解压 ZIP。
            </div>
          </div>
        ) : null}
        {step === 2 ? (
          <div className="grid gap-3 md:grid-cols-2">
            <Info label="runtime" value="python 3.12" />
            <Info label="entrypoint" value="src/main.py" />
            <Info label="auth_mode" value="reusable_session" />
            <Info label="trigger_type" value="manual / scheduled" />
            <Info label="域名白名单" value="archive.example.invalid" />
            <Info label="限制" value="512 MB / 600s / 512 MB 下载" />
          </div>
        ) : null}
        {step === 3 ? (
          <div className="grid gap-4">
            <div className="flex flex-wrap gap-2">
              <Badge tone="success">校验通过</Badge>
              <Badge tone="warning">权限待审核</Badge>
              <Badge>版本 1.5.0</Badge>
            </div>
            <div className="rounded-lg border border-warning/30 bg-warning/5 p-4">
              <div className="flex items-center gap-2 font-semibold">
                <ShieldAlert className="h-5 w-5 text-warning" /> 下一步建议
              </div>
              <p className="mt-2 text-sm text-secondary">审核域名白名单和资源限制，通过后才允许绑定到 Source。</p>
            </div>
            <Button icon={<CheckCircle2 className="h-4 w-4" />} onClick={finish}>登记模拟 Connector</Button>
          </div>
        ) : null}
      </section>
    </div>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border bg-subtle p-4">
      <p className="text-xs text-muted">{label}</p>
      <p className="mt-1 font-medium">{value}</p>
    </div>
  );
}

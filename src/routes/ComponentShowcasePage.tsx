import { Save } from "lucide-react";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Input, SearchBar, Select } from "../components/Form";
import { PageHeader, Section } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState, PermissionDeniedState, SuccessFeedback } from "../components/StateBlocks";

export function ComponentShowcasePage() {
  return (
    <div className="mx-auto max-w-6xl px-5 py-8">
      <PageHeader eyebrow="内部组件展示" title="共享组件与状态语言" />
      <Section title="控件">
        <div className="grid gap-4 rounded-lg border border-border bg-surface p-5 md:grid-cols-2">
          <Input label="输入框" placeholder="示例内容" />
          <Select label="选择器" options={["全部", "启用", "暂停"]} />
          <SearchBar />
          <div className="flex flex-wrap gap-2">
            <Button icon={<Save className="h-4 w-4" />}>主要操作</Button>
            <Button variant="secondary">次要操作</Button>
            <Button variant="ghost">弱操作</Button>
            <Button variant="danger">危险操作</Button>
          </div>
        </div>
      </Section>
      <Section title="标签">
        <div className="flex flex-wrap gap-2">
          <Badge tone="success">已通过</Badge>
          <Badge tone="warning">等待中</Badge>
          <Badge tone="danger">失败</Badge>
          <Badge tone="info">运行中</Badge>
          <Badge>草稿</Badge>
        </div>
      </Section>
      <Section title="状态">
        <div className="grid gap-4 md:grid-cols-2">
          <LoadingState />
          <EmptyState title="暂无记录" action="添加项目" />
          <ErrorState />
          <PermissionDeniedState />
          <SuccessFeedback message="静态状态已完成。" />
        </div>
      </Section>
    </div>
  );
}

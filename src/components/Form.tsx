import { Search } from "lucide-react";
import type { InputHTMLAttributes, SelectHTMLAttributes } from "react";

interface FieldProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string;
  hint?: string;
}

export function Input({ label, hint, className = "", ...props }: FieldProps) {
  return (
    <label className="grid gap-2 text-sm font-medium text-primary">
      {label}
      <input
        className={`min-h-11 rounded-md border border-border bg-surface px-3 text-primary placeholder:text-muted ${className}`}
        {...props}
      />
      {hint ? <span className="text-xs font-normal text-muted">{hint}</span> : null}
    </label>
  );
}

interface SearchBarProps extends InputHTMLAttributes<HTMLInputElement> {
  placeholder?: string;
}

export function SearchBar({ placeholder = "搜索", ...props }: SearchBarProps) {
  return (
    <label className="relative block">
      <span className="sr-only">搜索</span>
      <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted" />
      <input className="min-h-11 w-full rounded-md border border-border bg-surface pl-10 pr-3 text-sm" placeholder={placeholder} {...props} />
    </label>
  );
}

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label: string;
  options: string[];
}

export function Select({ label, options, className = "", ...props }: SelectProps) {
  return (
    <label className="grid gap-2 text-sm font-medium text-primary">
      {label}
      <select className={`min-h-11 rounded-md border border-border bg-surface px-3 text-sm ${className}`} {...props}>
        {options.map((option) => (
          <option key={option}>{option}</option>
        ))}
      </select>
    </label>
  );
}

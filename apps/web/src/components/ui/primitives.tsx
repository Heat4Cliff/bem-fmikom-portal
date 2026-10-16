import * as React from "react";
import Link from "next/link";
import * as AvatarPrimitive from "@radix-ui/react-avatar";
import * as LabelPrimitive from "@radix-ui/react-label";
import * as SeparatorPrimitive from "@radix-ui/react-separator";
import { AlertCircle, CheckCircle2, Info, Search, XCircle } from "lucide-react";
import { cn } from "@/lib/utils/cn";

export const Label = React.forwardRef<
  React.ElementRef<typeof LabelPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof LabelPrimitive.Root>
>(({ className, ...props }, ref) => (
  <LabelPrimitive.Root ref={ref} className={cn("text-sm font-semibold text-primary", className)} {...props} />
));
Label.displayName = "Label";

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  error?: string;
}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, error, id, ...props }, ref) => {
    const errorId = error && id ? `${id}-error` : undefined;
    return (
      <>
        <input
          ref={ref}
          id={id}
          className={cn(
            "flex h-11 w-full rounded-md border border-border bg-surface px-3 text-sm text-foreground shadow-sm placeholder:text-muted focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-secondary/20 disabled:cursor-not-allowed disabled:bg-surface-muted disabled:opacity-60 aria-[invalid=true]:border-danger",
            className,
          )}
          {...props}
          aria-invalid={error ? true : props["aria-invalid"]}
          aria-describedby={errorId ?? props["aria-describedby"]}
        />
        {error && errorId ? (
          <p id={errorId} className="mt-1.5 text-sm text-danger">
            {error}
          </p>
        ) : null}
      </>
    );
  },
);
Input.displayName = "Input";

export interface TextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  error?: string;
}

export const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, error, id, ...props }, ref) => {
    const errorId = error && id ? `${id}-error` : undefined;
    return (
      <>
        <textarea
          ref={ref}
          id={id}
          className={cn(
            "min-h-28 w-full rounded-md border border-border bg-surface px-3 py-2.5 text-sm text-foreground shadow-sm placeholder:text-muted focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-secondary/20 disabled:cursor-not-allowed disabled:opacity-60 aria-[invalid=true]:border-danger",
            className,
          )}
          {...props}
          aria-invalid={error ? true : props["aria-invalid"]}
          aria-describedby={errorId ?? props["aria-describedby"]}
        />
        {error && errorId ? (
          <p id={errorId} className="mt-1.5 text-sm text-danger">
            {error}
          </p>
        ) : null}
      </>
    );
  },
);
Textarea.displayName = "Textarea";

export function Select(props: React.SelectHTMLAttributes<HTMLSelectElement>) {
  const { className, children, ...rest } = props;
  return (
    <select
      className={cn(
        "h-11 w-full rounded-md border border-border bg-surface px-3 text-sm text-foreground shadow-sm focus-visible:border-primary disabled:cursor-not-allowed disabled:opacity-60",
        className,
      )}
      {...rest}
    >
      {children}
    </select>
  );
}

export function Checkbox({ className, ...props }: Omit<React.InputHTMLAttributes<HTMLInputElement>, "type">) {
  return (
    <input
      type="checkbox"
      className={cn(
        "size-4 rounded border-border accent-primary focus-visible:ring-2 focus-visible:ring-secondary/30 disabled:cursor-not-allowed",
        className,
      )}
      {...props}
    />
  );
}

export function RadioGroup({ className, ...props }: React.FieldsetHTMLAttributes<HTMLFieldSetElement>) {
  return <fieldset className={cn("flex flex-wrap gap-4", className)} {...props} />;
}

export function RadioOption({
  id,
  label,
  name,
  value,
  ...props
}: Omit<React.InputHTMLAttributes<HTMLInputElement>, "type"> & { label: string }) {
  return (
    <label htmlFor={id} className="inline-flex min-h-11 cursor-pointer items-center gap-2 text-sm text-body">
      <input id={id} name={name} value={value} type="radio" className="size-4 accent-primary" {...props} />
      <span>{label}</span>
    </label>
  );
}

const badgeStyles: Record<"neutral" | "info" | "success" | "warning" | "danger", string> = {
  neutral: "bg-surface-muted text-primary",
  info: "bg-info/10 text-info",
  success: "bg-success/10 text-success-foreground",
  warning: "bg-warning/15 text-warning-foreground",
  danger: "bg-danger/10 text-danger-foreground",
};

export function Badge({
  variant = "neutral",
  className,
  ...props
}: React.HTMLAttributes<HTMLSpanElement> & { variant?: keyof typeof badgeStyles }) {
  return (
    <span
      className={cn(
        "inline-flex min-h-6 items-center rounded-full px-2.5 py-1 text-xs font-semibold",
        badgeStyles[variant],
        className,
      )}
      {...props}
    />
  );
}

export function Card({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={cn("rounded-lg border border-border/70 bg-surface shadow-card", className)} {...props} />
  );
}

const alertTone = {
  info: { className: "border-info/25 bg-info/5 text-info", Icon: Info },
  success: { className: "border-success/25 bg-success/5 text-success-foreground", Icon: CheckCircle2 },
  warning: { className: "border-warning/30 bg-warning/10 text-warning-foreground", Icon: AlertCircle },
  danger: { className: "border-danger/25 bg-danger/5 text-danger-foreground", Icon: XCircle },
} as const;

export function Alert({
  variant = "info",
  title,
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLDivElement> & { variant?: keyof typeof alertTone; title?: string }) {
  const tone = alertTone[variant];
  const Icon = tone.Icon;
  return (
    <div
      role={variant === "danger" ? "alert" : "status"}
      className={cn("flex gap-3 rounded-md border p-4 text-sm", tone.className, className)}
      {...props}
    >
      <Icon aria-hidden="true" className="mt-0.5 size-4 shrink-0" />
      <div>
        {title ? <p className="font-semibold">{title}</p> : null}
        {children ? <div className={cn(title && "mt-1", "leading-6")}>{children}</div> : null}
      </div>
    </div>
  );
}

export function Separator({
  className,
  orientation = "horizontal",
  decorative = true,
  ...props
}: React.ComponentPropsWithoutRef<typeof SeparatorPrimitive.Root>) {
  return (
    <SeparatorPrimitive.Root
      decorative={decorative}
      orientation={orientation}
      className={cn(
        "shrink-0 bg-border",
        orientation === "horizontal" ? "h-px w-full" : "h-full w-px",
        className,
      )}
      {...props}
    />
  );
}

export const Avatar = React.forwardRef<
  React.ElementRef<typeof AvatarPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof AvatarPrimitive.Root>
>(({ className, ...props }, ref) => (
  <AvatarPrimitive.Root
    ref={ref}
    className={cn("relative flex size-10 shrink-0 overflow-hidden rounded-full bg-surface-soft", className)}
    {...props}
  />
));
Avatar.displayName = "Avatar";

export const AvatarImage = AvatarPrimitive.Image;
export const AvatarFallback = React.forwardRef<
  React.ElementRef<typeof AvatarPrimitive.Fallback>,
  React.ComponentPropsWithoutRef<typeof AvatarPrimitive.Fallback>
>(({ className, ...props }, ref) => (
  <AvatarPrimitive.Fallback
    ref={ref}
    className={cn(
      "grid size-full place-items-center bg-primary-container text-xs font-bold text-primary-container-foreground",
      className,
    )}
    {...props}
  />
));
AvatarFallback.displayName = "AvatarFallback";

export function Breadcrumb({ children, className, ...props }: React.HTMLAttributes<HTMLElement>) {
  return (
    <nav aria-label="Breadcrumb" className={cn("text-sm text-muted", className)} {...props}>
      <ol className="flex flex-wrap items-center gap-2">{children}</ol>
    </nav>
  );
}

export function BreadcrumbItem({
  current = false,
  children,
}: {
  current?: boolean;
  children: React.ReactNode;
}) {
  return (
    <li
      aria-current={current ? "page" : undefined}
      className={cn("inline-flex items-center gap-2", current && "font-semibold text-primary")}
    >
      {children}
    </li>
  );
}

export function PageHeader({
  title,
  description,
  actions,
  className,
}: {
  title: string;
  description?: string;
  actions?: React.ReactNode;
  className?: string;
}) {
  return (
    <header
      className={cn(
        "mb-6 flex min-w-0 flex-col gap-4 md:flex-row md:items-end md:justify-between",
        className,
      )}
    >
      <div className="min-w-0">
        <h1 className="font-heading text-2xl font-bold tracking-tight text-primary md:text-3xl">{title}</h1>
        {description ? <p className="mt-1.5 max-w-2xl text-sm leading-6 text-muted">{description}</p> : null}
      </div>
      {actions ? (
        <div
          data-page-header-actions
          className="flex w-full min-w-0 flex-col items-stretch gap-3 md:w-auto md:flex-row md:flex-wrap md:items-center md:justify-end [&>*]:w-full [&>*]:justify-center [&>*]:whitespace-nowrap [&_a]:w-full [&_button]:w-full [&_select]:w-full md:[&>*]:w-auto md:[&_a]:w-auto md:[&_button]:w-auto md:[&_select]:w-auto"
        >
          {actions}
        </div>
      ) : null}
    </header>
  );
}

export function ActionGroup({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        "flex min-w-0 flex-col items-stretch gap-3 sm:flex-row sm:flex-wrap sm:items-center",
        className,
      )}
      {...props}
    />
  );
}

export function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      aria-hidden="true"
      className={cn("animate-pulse rounded-md bg-surface-raised", className)}
      {...props}
    />
  );
}

export function Spinner({ label = "Memuat", className }: { label?: string; className?: string }) {
  return (
    <span role="status" className={cn("inline-flex items-center gap-2 text-sm text-muted", className)}>
      <span
        aria-hidden="true"
        className="size-4 animate-spin rounded-full border-2 border-primary/20 border-t-primary"
      />
      <span className="sr-only">{label}</span>
    </span>
  );
}

export function EmptyState({
  title,
  description,
  action,
  className,
}: {
  title: string;
  description?: string;
  action?: React.ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center rounded-lg border border-dashed border-border bg-surface px-6 py-12 text-center",
        className,
      )}
    >
      <p className="font-heading font-semibold text-primary">{title}</p>
      {description ? <p className="mt-2 max-w-md text-sm leading-6 text-muted">{description}</p> : null}
      {action ? <div className="mt-5">{action}</div> : null}
    </div>
  );
}

export function ErrorState({
  title = "Terjadi kendala",
  description,
  retry,
  className,
}: {
  title?: string;
  description?: string;
  retry?: React.ReactNode;
  className?: string;
}) {
  return (
    <div role="alert" className={cn("rounded-lg border border-danger/25 bg-danger/5 p-5", className)}>
      <p className="font-semibold text-danger-foreground">{title}</p>
      {description ? <p className="mt-1 text-sm leading-6 text-danger-foreground">{description}</p> : null}
      {retry ? <div className="mt-4">{retry}</div> : null}
    </div>
  );
}

export function Pagination({
  page,
  totalPages,
  baseHref = "?page=",
}: {
  page: number;
  totalPages: number;
  baseHref?: string;
}) {
  const previousDisabled = page <= 1;
  const nextDisabled = page >= totalPages;
  return (
    <nav aria-label="Pagination" className="flex items-center justify-between gap-3 text-sm">
      <span className="text-muted">
        Halaman {page} dari {Math.max(totalPages, 1)}
      </span>
      <div className="flex gap-2">
        <Link
          aria-disabled={previousDisabled}
          tabIndex={previousDisabled ? -1 : undefined}
          className="rounded-md border border-border px-3 py-2 text-primary aria-disabled:opacity-50"
          href={`${baseHref}${Math.max(1, page - 1)}`}
        >
          Sebelumnya
        </Link>
        <Link
          aria-disabled={nextDisabled}
          tabIndex={nextDisabled ? -1 : undefined}
          className="rounded-md border border-border px-3 py-2 text-primary aria-disabled:opacity-50"
          href={`${baseHref}${Math.min(totalPages, page + 1)}`}
        >
          Berikutnya
        </Link>
      </div>
    </nav>
  );
}

export function SearchInput({
  label = "Cari",
  className,
  ...props
}: React.InputHTMLAttributes<HTMLInputElement> & { label?: string }) {
  return (
    <div className={cn("relative block", className)}>
      <span className="sr-only">{label}</span>
      <Search
        aria-hidden="true"
        className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted"
      />
      <Input className="pl-9" type="search" aria-label={label} {...props} />
    </div>
  );
}

export function FilterTrigger({
  children,
  className,
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      type="button"
      className={cn(
        "inline-flex min-h-10 items-center gap-2 rounded-md border border-border bg-surface px-3 text-sm font-medium text-primary hover:bg-surface-muted",
        className,
      )}
      {...props}
    >
      {children}
    </button>
  );
}

export function FileInput({
  id,
  label,
  hint,
  className,
  ...props
}: React.InputHTMLAttributes<HTMLInputElement> & { label: string; hint?: string }) {
  return (
    <div className={cn("grid gap-2", className)}>
      <Label htmlFor={id}>{label}</Label>
      <input
        id={id}
        type="file"
        className="block w-full cursor-pointer rounded-md border border-border bg-surface text-sm text-body file:mr-4 file:min-h-11 file:border-0 file:bg-surface-muted file:px-4 file:font-semibold file:text-primary hover:file:bg-surface-soft"
        {...props}
      />
      {hint ? <p className="text-xs text-muted">{hint}</p> : null}
    </div>
  );
}

export function Table({ className, ...props }: React.TableHTMLAttributes<HTMLTableElement>) {
  return (
    <div className="w-full overflow-hidden rounded-lg border border-border bg-surface">
      <div className="overflow-x-auto">
        <table className={cn("w-full caption-bottom text-sm", className)} {...props} />
      </div>
    </div>
  );
}

export function TableHeadCell({ className, ...props }: React.ThHTMLAttributes<HTMLTableCellElement>) {
  return (
    <th
      scope="col"
      className={cn(
        "h-12 bg-surface-muted px-4 text-left text-xs font-bold uppercase tracking-wide text-muted",
        className,
      )}
      {...props}
    />
  );
}

export function TableCell({ className, ...props }: React.TdHTMLAttributes<HTMLTableCellElement>) {
  return (
    <td className={cn("border-t border-border/70 px-4 py-3 align-middle text-body", className)} {...props} />
  );
}

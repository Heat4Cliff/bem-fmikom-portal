import type { ReactNode } from "react";

export function SectionHeading({
  eyebrow,
  title,
  description,
  action,
  centered = false,
}: {
  eyebrow: string;
  title: string;
  description?: string;
  action?: ReactNode;
  centered?: boolean;
}) {
  return (
    <div
      className={`mb-7 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between ${centered ? "text-center sm:flex-col sm:items-center" : ""}`}
    >
      <div className="max-w-2xl">
        <p className="text-xs font-bold uppercase tracking-[0.16em] text-secondary">{eyebrow}</p>
        <h2 className="mt-2 font-heading text-2xl font-bold leading-tight tracking-tight text-primary md:text-3xl">
          {title}
        </h2>
        {description ? (
          <p className="mt-3 max-w-xl text-sm leading-6 text-muted md:text-base">{description}</p>
        ) : null}
      </div>
      {action ? <div className="shrink-0">{action}</div> : null}
    </div>
  );
}

export function EmptyPreview({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="rounded-xl border border-dashed border-border bg-white/70 px-5 py-7 text-center sm:px-8">
      <p className="font-heading text-base font-semibold text-primary">{title}</p>
      <p className="mx-auto mt-2 max-w-xl text-sm leading-6 text-muted">{children}</p>
    </div>
  );
}

export function DemoDataNote() {
  return (
    <p className="mb-4 inline-flex rounded-full bg-amber-100 px-3 py-1 text-xs font-semibold text-amber-950">
      Contoh visual · bukan data resmi
    </p>
  );
}

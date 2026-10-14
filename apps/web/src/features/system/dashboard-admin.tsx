"use client";

import * as React from "react";
import Link from "next/link";
import { Archive, FileText, Megaphone, MessageSquareText, UserRoundPlus } from "lucide-react";
import { AdminPageHeader } from "@/components/admin/admin-page-header";
import { Alert, Card } from "@/components/ui/primitives";
import { CountUp } from "@/components/react-bits/count-up";
import { getDashboard, type DashboardSummary } from "./api";

function Metric({
  title,
  value,
  detail,
  icon: Icon,
  href,
}: {
  title: string;
  value: number;
  detail?: string;
  icon: typeof FileText;
  href: string;
}) {
  return (
    <Card className="p-4 sm:p-5">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-xs font-semibold uppercase tracking-wide text-muted">{title}</p>
          <p className="mt-2 text-3xl font-bold tabular-nums text-primary">
            <CountUp to={value} />
          </p>
          {detail ? <p className="mt-2 text-sm text-muted">{detail}</p> : null}
        </div>
        <span className="grid size-10 shrink-0 place-items-center rounded-xl bg-primary/5 text-primary">
          <Icon aria-hidden="true" className="size-5" />
        </span>
      </div>
      <Link
        href={href}
        className="mt-4 inline-flex min-h-10 items-center text-sm font-semibold text-primary underline-offset-4 hover:underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2"
      >
        Buka modul <span className="sr-only">{title}</span>
      </Link>
    </Card>
  );
}

export function DashboardAdmin() {
  const [summary, setSummary] = React.useState<DashboardSummary | null>(null);
  const [error, setError] = React.useState(false);
  React.useEffect(() => {
    let mounted = true;
    void getDashboard()
      .then((data) => {
        if (mounted) setSummary(data);
      })
      .catch(() => {
        if (mounted) setError(true);
      });
    return () => {
      mounted = false;
    };
  }, []);

  if (error)
    return (
      <div className="mx-auto w-full max-w-[1440px] p-4 md:p-6 xl:p-8">
        <AdminPageHeader title="Dashboard" />
        <Alert className="mt-6" variant="danger" title="Ringkasan belum dapat dimuat">
          Coba muat ulang halaman. Data hanya menampilkan ringkasan sesuai kewenangan akun.
        </Alert>
      </div>
    );
  if (!summary)
    return (
      <div className="mx-auto w-full max-w-[1440px] p-4 md:p-6 xl:p-8">
        <AdminPageHeader title="Dashboard" />
        <section
          className="mt-6 grid gap-4 sm:grid-cols-2 xl:grid-cols-5"
          aria-busy="true"
          aria-label="Memuat ringkasan dashboard"
          role="status"
        >
          {["article", "registrations", "aspirations", "documents", "announcements"].map((key) => (
            <Card key={key} className="h-36 animate-pulse bg-surface-muted" />
          ))}
        </section>
      </div>
    );

  const period = summary.active_period;
  const scope =
    summary.scope.type === "division" ? `Divisi ${summary.scope.division?.name ?? ""}` : "Seluruh organisasi";
  return (
    <div className="mx-auto w-full max-w-[1440px] p-4 md:p-6 xl:p-8">
      <AdminPageHeader title="Dashboard" description={`Ringkasan operasional · ${scope}`} />
      <Card className="mt-5 flex flex-wrap items-center justify-between gap-3 p-4">
        <div>
          <p className="text-xs font-semibold uppercase tracking-wide text-muted">Periode aktif</p>
          <p className="mt-1 font-semibold text-primary">{period?.name ?? "Belum ada periode aktif"}</p>
        </div>
        <p className="text-xs text-muted">
          Diperbarui{" "}
          {new Date(summary.generated_at).toLocaleString("id-ID", {
            dateStyle: "medium",
            timeStyle: "short",
          })}
        </p>
      </Card>
      <section aria-label="Statistik operasional" className="mt-5 grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
        <Metric
          title="Berita"
          value={summary.articles.total}
          detail={`${summary.articles.published} terpublikasi · ${summary.articles.draft} draft`}
          icon={FileText}
          href="/admin/berita"
        />
        <Metric
          title="Pendaftaran aktif"
          value={summary.registrations.active_forms}
          detail={`${summary.registrations.active_form_submissions.toLocaleString("id-ID")} pendaftar pada formulir aktif`}
          icon={UserRoundPlus}
          href="/admin/pendaftaran"
        />
        <Metric
          title="Aspirasi masuk"
          value={summary.aspirations.total}
          detail={`${summary.aspirations.pending} menunggu · ${summary.aspirations.read} dibaca · ${summary.aspirations.resolved} selesai`}
          icon={MessageSquareText}
          href="/admin/aspirasi"
        />
        <Metric
          title="Arsip dokumen"
          value={summary.documents.total}
          detail="Dokumen yang masih tersedia"
          icon={Archive}
          href="/admin/arsip-dokumen"
        />
        <Metric
          title="Pengumuman aktif"
          value={summary.announcements.total}
          detail="Dalam jadwal tayang saat ini"
          icon={Megaphone}
          href="/admin/pengumuman"
        />
      </section>
      <p className="mt-5 text-xs leading-5 text-muted">
        Ringkasan tidak mengumpulkan kunjungan atau data pelacakan pengunjung. Jumlah artikel, formulir,
        aspirasi, dan dokumen mengikuti scope akses akun.
      </p>
    </div>
  );
}

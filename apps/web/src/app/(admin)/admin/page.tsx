import { LayoutDashboard } from "lucide-react";
import { AdminPageHeader } from "@/components/admin/admin-page-header";
import { Alert, Card } from "@/components/ui/primitives";

export default function AdminFoundationPage() {
  return (
    <div className="mx-auto w-full max-w-[1440px] p-4 md:p-6 xl:p-8">
      <AdminPageHeader title="Dashboard" description="Fondasi dashboard admin BEM FMIKOM UNUGHA." />
      <Card className="mt-6 flex min-h-72 flex-col items-center justify-center border-primary/10 px-6 py-12 text-center">
        <span className="mb-4 grid size-14 place-items-center rounded-2xl bg-surface-soft text-primary">
          <LayoutDashboard aria-hidden="true" className="size-7" />
        </span>
        <h2 className="font-heading text-xl font-bold text-primary">Kerangka admin siap</h2>
        <p className="mt-2 max-w-lg text-sm leading-6 text-muted">
          Sesi administrator telah diverifikasi. Ringkasan data organisasi akan tersedia setelah modul
          dashboard masuk tahap implementasi.
        </p>
        <Alert className="mt-6 max-w-xl text-left" variant="info" title="Fondasi Phase 2">
          Login, session refresh, dan navigasi berdasarkan role sudah aktif. Data bisnis belum dimuat pada
          tahap ini.
        </Alert>
      </Card>
    </div>
  );
}

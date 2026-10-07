import { ArrowRight, Sparkles } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/primitives";

export default function PublicFoundationPage() {
  return (
    <section className="mx-auto grid min-h-[64vh] w-full max-w-[var(--container-public)] content-center gap-8 px-4 py-16 md:grid-cols-[1.15fr_0.85fr] md:px-8 lg:px-12">
      <div className="flex flex-col items-start justify-center">
        <span className="mb-5 inline-flex items-center gap-2 rounded-full border border-border bg-surface px-3 py-1.5 text-xs font-semibold uppercase tracking-wide text-primary">
          <Sparkles aria-hidden="true" className="size-3.5 text-secondary" />
          Portal resmi BEM FMIKOM UNUGHA
        </span>
        <h1 className="max-w-3xl font-heading text-4xl font-extrabold leading-tight tracking-tight text-primary md:text-5xl lg:text-6xl">
          Ruang bersama untuk bertumbuh dan berkarya.
        </h1>
        <p className="mt-5 max-w-xl text-base leading-7 text-muted md:text-lg">
          Fondasi portal publik dan sistem pengelolaan organisasi. Konten dan layanan BEM akan hadir pada fase
          berikutnya.
        </p>
        <div className="mt-8 flex flex-wrap gap-3">
          <Button asChild>
            <Link href="/berita">
              Jelajahi portal <ArrowRight aria-hidden="true" className="size-4" />
            </Link>
          </Button>
          <Button asChild variant="outline">
            <Link href="/aspirasi">Kirim aspirasi</Link>
          </Button>
        </div>
      </div>
      <Card className="flex min-h-64 flex-col justify-between overflow-hidden border-primary/10 bg-gradient-to-br from-white via-surface-muted to-surface-soft p-6 md:min-h-80 md:p-8">
        <span className="text-xs font-bold uppercase tracking-[0.18em] text-secondary">
          Kabinet Sinergi Progresif
        </span>
        <div>
          <div className="mb-4 h-1 w-16 rounded-full bg-secondary" />
          <p className="font-heading text-2xl font-bold leading-snug text-primary md:text-3xl">
            Informasi organisasi dalam satu portal.
          </p>
          <p className="mt-3 max-w-sm text-sm leading-6 text-muted">
            Kerangka navigasi publik ini menyiapkan tempat bagi berita, kegiatan, pengumuman, pendaftaran, dan
            aspirasi.
          </p>
        </div>
      </Card>
    </section>
  );
}

import type { Metadata } from "next";
import { Check, Clock3, LockKeyhole, MessageSquareHeart } from "lucide-react";
import { PublicAspirationForm } from "@/features/aspirations/public-aspiration-form";

export const metadata: Metadata = {
  title: "Aspirasi Mahasiswa",
  description: "Sampaikan aspirasi kepada BEM FMIKOM UNUGHA tanpa perlu membuat akun.",
};

export default function AspirationPage() {
  return (
    <main className="bg-background">
      <section className="mx-auto w-full max-w-[var(--container-public)] px-4 py-10 md:px-8 md:py-14 lg:px-12 lg:py-16">
        <div className="grid items-start gap-8 lg:grid-cols-[minmax(0,1.3fr)_minmax(18rem,0.7fr)] lg:gap-10">
          <div>
            <p className="inline-flex items-center gap-2 rounded-full bg-primary/10 px-3 py-1.5 text-xs font-semibold text-primary">
              <span className="size-2 rounded-full bg-secondary" /> Kanal aspirasi resmi BEM FMIKOM UNUGHA
            </p>
            <h1 className="mt-5 max-w-3xl font-heading text-3xl font-bold leading-tight tracking-tight text-primary sm:text-4xl">
              Kotak Aspirasi & Advokasi Mahasiswa FMIKOM
            </h1>
            <p className="mt-4 max-w-2xl text-sm leading-7 text-body sm:text-base">
              Sampaikan saran, keluhan, atau ide kepada pengurus BEM. Pilih tujuan agar aspirasi diteruskan
              kepada divisi yang relevan.
            </p>

            <div className="mt-8">
              <PublicAspirationForm />
            </div>
          </div>

          <aside className="space-y-4 lg:sticky lg:top-24">
            <section className="rounded-2xl border border-border bg-white p-5 shadow-card sm:p-6">
              <span className="grid size-10 place-items-center rounded-lg bg-primary/10 text-primary">
                <LockKeyhole aria-hidden="true" className="size-5" />
              </span>
              <h2 className="mt-4 font-heading text-lg font-bold text-primary">Aman dan terjaga</h2>
              <p className="mt-2 text-sm leading-6 text-body">
                Pengiriman tidak memerlukan akun. Identitas bersifat opsional dan dapat dihapus dari
                pengiriman dengan memilih anonim.
              </p>
            </section>

            <section className="rounded-2xl border border-border bg-white p-5 shadow-card sm:p-6">
              <h2 className="font-heading text-lg font-bold text-primary">Alur aspirasi</h2>
              <ol className="mt-5 space-y-4">
                {[
                  {
                    title: "Dikirim",
                    description: "Aspirasi diterima oleh sistem.",
                    icon: MessageSquareHeart,
                  },
                  { title: "Dibaca", description: "Pengurus divisi tujuan meninjau masukan.", icon: Clock3 },
                  { title: "Selesai", description: "Status diperbarui setelah tindak lanjut.", icon: Check },
                ].map(({ title, description, icon: Icon }, index) => (
                  <li key={title} className="flex gap-3">
                    <span className="grid size-8 shrink-0 place-items-center rounded-full bg-surface-muted text-primary">
                      <Icon aria-hidden="true" className="size-4" />
                    </span>
                    <span>
                      <strong className="block text-sm text-primary">
                        {index + 1}. {title}
                      </strong>
                      <span className="mt-0.5 block text-xs leading-5 text-muted">{description}</span>
                    </span>
                  </li>
                ))}
              </ol>
            </section>
          </aside>
        </div>
      </section>
    </main>
  );
}

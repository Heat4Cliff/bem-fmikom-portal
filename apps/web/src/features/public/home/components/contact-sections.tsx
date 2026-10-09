import { ArrowRight, Mail, MapPin, MessageSquareHeart, Phone } from "lucide-react";
import Link from "next/link";
import type { ContactInfo } from "../types";
import { Button } from "@/components/ui/button";
import { EmptyPreview, SectionHeading } from "./section-primitives";

export function AspirationSection() {
  return (
    <section
      id="aspirasi"
      className="scroll-mt-24 bg-primary px-4 py-14 text-white md:px-8 lg:px-12 lg:py-20"
    >
      <div className="mx-auto grid max-w-[var(--container-public)] items-center gap-7 lg:grid-cols-[1fr_auto]">
        <div className="max-w-2xl">
          <span className="grid size-11 place-items-center rounded-xl bg-amber-400 text-amber-950">
            <MessageSquareHeart aria-hidden="true" className="size-5" />
          </span>
          <p className="mt-4 text-xs font-bold uppercase tracking-[0.16em] text-amber-200">Suara mahasiswa</p>
          <h2 className="mt-2 font-heading text-2xl font-bold leading-tight sm:text-3xl">
            Aspirasi Anda membangun fakultas lebih baik
          </h2>
          <p className="mt-3 text-sm leading-6 text-blue-100 sm:text-base">
            Ruang aspirasi disiapkan sebagai kanal partisipasi. Formulir dan pengiriman aspirasi belum
            tersedia pada tahap ini.
          </p>
        </div>
        <Button asChild size="lg" variant="secondary" className="w-full sm:w-fit">
          <Link href="/aspirasi">
            Lihat informasi aspirasi <ArrowRight aria-hidden="true" className="size-4" />
          </Link>
        </Button>
      </div>
    </section>
  );
}

export function ContactSection({ contact, demoMode }: { contact: ContactInfo; demoMode: boolean }) {
  const entries = [
    { icon: MapPin, label: "Alamat", value: contact.address },
    { icon: Mail, label: "Email", value: contact.email },
    { icon: Phone, label: "Telepon", value: contact.phone },
  ].filter((item) => item.value);

  return (
    <section id="kontak" className="scroll-mt-24 bg-background px-4 py-14 md:px-8 lg:px-12 lg:py-20">
      <div className="mx-auto max-w-[var(--container-public)]">
        <SectionHeading
          eyebrow="Tetap terhubung"
          title="Kontak & Media Sosial"
          description="Kanal resmi akan ditambahkan setelah alamat kontak dan akun sosial organisasi dikonfirmasi."
        />
        {demoMode ? (
          <p className="mb-4 text-xs font-semibold text-amber-900">
            Contoh visual · kontak resmi belum dikonfirmasi
          </p>
        ) : null}
        {entries.length || contact.socialLinks.length ? (
          <div className="grid gap-5 rounded-2xl border border-border/70 bg-white p-5 shadow-card sm:grid-cols-2 sm:p-7">
            <div className="space-y-4">
              {entries.map(({ icon: Icon, label, value }) => (
                <div key={label} className="flex items-start gap-3">
                  <span className="grid size-9 shrink-0 place-items-center rounded-lg bg-surface-muted text-primary">
                    <Icon aria-hidden="true" className="size-4" />
                  </span>
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-wide text-muted">{label}</p>
                    <p className="mt-1 break-words text-sm font-medium text-primary">{value}</p>
                  </div>
                </div>
              ))}
            </div>
            {contact.socialLinks.length ? (
              <nav aria-label="Media sosial BEM FMIKOM" className="flex flex-wrap content-start gap-2">
                {contact.socialLinks.map((social) => (
                  <a
                    key={social.label}
                    href={social.href}
                    target="_blank"
                    rel="noreferrer"
                    className="rounded-full border border-border px-4 py-2 text-sm font-semibold text-primary hover:bg-surface-muted focus-visible:outline-2 focus-visible:outline-secondary"
                  >
                    {social.label}
                  </a>
                ))}
              </nav>
            ) : (
              <EmptyPreview title="Media sosial resmi belum tersedia">
                Tidak ada tautan sosial contoh yang ditampilkan di mode produksi.
              </EmptyPreview>
            )}
          </div>
        ) : (
          <EmptyPreview title="Kontak resmi belum tersedia">
            Informasi alamat, email, telepon, dan akun media sosial akan ditambahkan setelah dikonfirmasi oleh
            organisasi.
          </EmptyPreview>
        )}
      </div>
    </section>
  );
}

"use client";

import * as React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  BookOpenText,
  CalendarDays,
  Camera,
  Home,
  Menu,
  MessageSquareText,
  Megaphone,
  Network,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Drawer } from "@/components/ui/overlays";
import { cn } from "@/lib/utils/cn";

const primaryLinks = [
  { label: "Beranda", href: "/", icon: Home },
  { label: "Tentang", href: "/#tentang" },
  { label: "Struktur", href: "/struktur", icon: Network },
  { label: "Berita", href: "/#berita", icon: BookOpenText },
  { label: "Galeri", href: "/#galeri", icon: Camera },
  { label: "Pengumuman", href: "/#pengumuman", icon: Megaphone },
  { label: "Pendaftaran", href: "/#pendaftaran", icon: CalendarDays },
  { label: "Aspirasi", href: "/#aspirasi", icon: MessageSquareText },
] as const;

function isCurrent(pathname: string, href: string) {
  const path = href.split("#")[0];
  return path === "/" ? pathname === "/" : pathname === path || pathname.startsWith(`${path}/`);
}

export function PublicNavigation({
  onNavigate,
  mobile = false,
}: {
  onNavigate?: () => void;
  mobile?: boolean;
}) {
  const pathname = usePathname();
  return (
    <nav aria-label="Navigasi utama">
      <ul className={mobile ? "flex flex-col gap-1" : "flex items-center gap-0.5"}>
        {primaryLinks.map(({ label, href }) => (
          <li key={label}>
            <Link
              onClick={onNavigate}
              aria-current={isCurrent(pathname, href) && !href.includes("#") ? "page" : undefined}
              className={cn(
                "block rounded-md px-2 py-2 text-xs font-semibold text-body transition-colors hover:bg-surface-muted hover:text-primary focus-visible:outline-2 focus-visible:outline-secondary aria-[current=page]:bg-surface-muted aria-[current=page]:text-primary xl:px-2.5 xl:text-sm",
                mobile && "px-3 py-3 text-sm",
                label === "Aspirasi" && "text-secondary",
              )}
              href={href}
            >
              {label}
            </Link>
          </li>
        ))}
      </ul>
    </nav>
  );
}

export function PublicHeader() {
  const [open, setOpen] = React.useState(false);
  return (
    <header className="sticky top-0 z-40 border-b border-border/70 bg-background/95 backdrop-blur">
      <div className="mx-auto flex h-[4.5rem] w-full max-w-[var(--container-public)] items-center justify-between gap-3 px-4 md:px-8 lg:px-12">
        <Link
          href="/"
          className="flex min-w-0 items-center gap-2.5 rounded-md"
          aria-label="BEM FMIKOM UNUGHA, beranda"
        >
          <span className="grid size-9 shrink-0 place-items-center rounded-lg bg-primary font-heading text-xs font-extrabold text-primary-foreground">
            BEM
          </span>
          <span className="flex min-w-0 flex-col leading-tight">
            <span className="truncate font-heading text-sm font-bold text-primary">BEM FMIKOM</span>
            <span className="text-[10px] font-medium tracking-wide text-muted">UNUGHA CILACAP</span>
          </span>
        </Link>
        <div className="hidden flex-1 justify-end xl:flex">
          <PublicNavigation />
        </div>
        <div className="flex items-center gap-2 xl:hidden">
          <Button
            aria-label="Buka menu navigasi"
            aria-haspopup="dialog"
            variant="outline"
            size="icon"
            onClick={() => setOpen(true)}
          >
            <Menu aria-hidden="true" className="size-5" />
          </Button>
        </div>
      </div>
      <Drawer
        open={open}
        onOpenChange={setOpen}
        title="Navigasi BEM FMIKOM"
        side="right"
        className="w-[min(88vw,24rem)]"
      >
        <div className="p-4">
          <PublicNavigation mobile onNavigate={() => setOpen(false)} />
        </div>
      </Drawer>
    </header>
  );
}

const mobileLinks = [
  { label: "Beranda", href: "/", icon: Home },
  { label: "Berita", href: "/#berita", icon: BookOpenText },
  { label: "Daftar", href: "/#pendaftaran", icon: CalendarDays },
  { label: "Aspirasi", href: "/aspirasi", icon: MessageSquareText },
  { label: "Profil", href: "/#tentang", icon: Network },
] as const;

export function MobileNavigation() {
  const pathname = usePathname();
  return (
    <nav
      aria-label="Navigasi bawah"
      className="fixed inset-x-0 bottom-0 z-40 border-t border-border bg-surface/95 px-2 pb-[max(env(safe-area-inset-bottom),0.35rem)] pt-1.5 backdrop-blur md:hidden"
    >
      <ul className="mx-auto grid max-w-md grid-cols-5">
        {mobileLinks.map(({ label, href, icon: Icon }) => {
          const active = isCurrent(pathname, href);
          return (
            <li key={label}>
              <Link
                aria-current={active ? "page" : undefined}
                href={href}
                className={cn(
                  "flex min-h-12 flex-col items-center justify-center gap-1 rounded-md text-[10px] font-semibold text-muted aria-[current=page]:text-primary",
                )}
              >
                <Icon aria-hidden="true" className={cn("size-[18px]", active && "stroke-[2.5]")} />
                <span>{label}</span>
              </Link>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}

export function PublicFooter() {
  return (
    <footer className="border-t border-border/70 bg-surface-soft px-4 pb-24 pt-10 md:px-8 md:pb-10 lg:px-12">
      <div className="mx-auto grid max-w-[var(--container-public)] gap-8 sm:grid-cols-2 lg:grid-cols-[1.4fr_1fr_1fr]">
        <div>
          <Link
            href="/"
            className="flex w-fit items-center gap-2.5 rounded-md focus-visible:outline-2 focus-visible:outline-secondary"
          >
            <span className="grid size-10 place-items-center rounded-lg bg-primary font-heading text-xs font-extrabold text-white">
              BEM
            </span>
            <span>
              <strong className="block font-heading text-sm font-bold text-primary">BEM FMIKOM</strong>
              <span className="text-[10px] font-medium tracking-wide text-muted">UNUGHA CILACAP</span>
            </span>
          </Link>
          <p className="mt-4 max-w-sm text-sm leading-6 text-muted">
            Portal publik informasi dan kegiatan BEM Fakultas Matematika dan Ilmu Komputer UNUGHA.
          </p>
        </div>
        <div>
          <h2 className="font-heading text-sm font-bold text-primary">Tautan cepat</h2>
          <ul className="mt-3 grid grid-cols-2 gap-x-3 gap-y-2 text-sm text-muted">
            <li>
              <Link className="hover:text-primary" href="/#tentang">
                Tentang
              </Link>
            </li>
            <li>
              <Link className="hover:text-primary" href="/struktur">
                Struktur
              </Link>
            </li>
            <li>
              <Link className="hover:text-primary" href="/#berita">
                Berita
              </Link>
            </li>
            <li>
              <Link className="hover:text-primary" href="/#galeri">
                Galeri
              </Link>
            </li>
            <li>
              <Link className="hover:text-primary" href="/#pengumuman">
                Pengumuman
              </Link>
            </li>
            <li>
              <Link className="hover:text-primary" href="/aspirasi">
                Aspirasi
              </Link>
            </li>
          </ul>
        </div>
        <div>
          <h2 className="font-heading text-sm font-bold text-primary">Sekretariat & kontak</h2>
          <p className="mt-3 text-sm leading-6 text-muted">
            Informasi kontak resmi akan ditambahkan setelah dikonfirmasi.
          </p>
          <p className="mt-6 text-xs text-muted">© {new Date().getFullYear()} BEM FMIKOM UNUGHA</p>
        </div>
      </div>
    </footer>
  );
}

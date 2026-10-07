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
  { label: "Struktur", href: "/#struktur", icon: Network },
  { label: "Berita", href: "/berita", icon: BookOpenText },
  { label: "Galeri", href: "/galeri", icon: Camera },
  { label: "Pengumuman", href: "/pengumuman", icon: Megaphone },
  { label: "Pendaftaran", href: "/pendaftaran", icon: CalendarDays },
  { label: "Aspirasi", href: "/aspirasi", icon: MessageSquareText },
] as const;

function isCurrent(pathname: string, href: string) {
  const path = href.split("#")[0];
  return path === "/" ? pathname === "/" : pathname === path || pathname.startsWith(`${path}/`);
}

export function PublicNavigation({ onNavigate }: { onNavigate?: () => void }) {
  const pathname = usePathname();
  return (
    <nav aria-label="Navigasi utama">
      <ul className="flex flex-col gap-1 md:flex-row md:items-center md:gap-0.5 lg:gap-1">
        {primaryLinks.map(({ label, href }) => (
          <li key={label}>
            <Link
              onClick={onNavigate}
              aria-current={isCurrent(pathname, href) ? "page" : undefined}
              className={cn(
                "block rounded-md px-3 py-2 text-sm font-medium text-body transition-colors hover:bg-surface-muted hover:text-primary aria-[current=page]:bg-surface-muted aria-[current=page]:font-semibold aria-[current=page]:text-primary md:px-2.5 lg:px-3",
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
      <div className="mx-auto flex h-[4.5rem] w-full max-w-[var(--container-public)] items-center justify-between gap-4 px-4 md:px-8 lg:px-12">
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
        <div className="hidden flex-1 justify-end md:flex">
          <PublicNavigation />
        </div>
        <div className="flex items-center gap-2 md:hidden">
          <Button aria-label="Buka menu navigasi" variant="outline" size="icon" onClick={() => setOpen(true)}>
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
          <PublicNavigation onNavigate={() => setOpen(false)} />
        </div>
        <div className="mt-auto border-t border-border p-4">
          <Link
            onClick={() => setOpen(false)}
            className="inline-flex min-h-11 w-full items-center justify-center gap-2 rounded-md bg-secondary px-4 text-sm font-semibold text-white"
            href="/aspirasi"
          >
            <MessageSquareText aria-hidden="true" className="size-4" />
            Kirim Aspirasi
          </Link>
        </div>
      </Drawer>
    </header>
  );
}

const mobileLinks = [
  { label: "Beranda", href: "/", icon: Home },
  { label: "Berita", href: "/berita", icon: BookOpenText },
  { label: "Daftar", href: "/pendaftaran", icon: CalendarDays },
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
    <footer className="border-t border-border/70 bg-surface-muted px-4 py-8 md:px-8">
      <div className="mx-auto flex max-w-[var(--container-public)] flex-col gap-3 text-sm text-muted sm:flex-row sm:items-center sm:justify-between lg:px-4">
        <div className="flex items-center gap-2">
          <span className="grid size-8 place-items-center rounded-md bg-primary text-[10px] font-bold text-white">
            BEM
          </span>
          <span>
            <strong className="font-semibold text-primary">BEM FMIKOM UNUGHA</strong>
            <br />
            <span className="text-xs">Portal organisasi mahasiswa</span>
          </span>
        </div>
        <p className="text-xs">© {new Date().getFullYear()} BEM FMIKOM UNUGHA</p>
      </div>
    </footer>
  );
}

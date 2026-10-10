"use client";

import {
  Activity,
  Bell,
  BookOpenText,
  CalendarDays,
  FileText,
  FolderOpen,
  GalleryVerticalEnd,
  LayoutDashboard,
  LogOut,
  Megaphone,
  Menu,
  MessageSquareText,
  Network,
  Settings2,
  ShieldCheck,
  Users,
  UsersRound,
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import * as React from "react";
import { Button } from "@/components/ui/button";
import { Drawer } from "@/components/ui/overlays";
import { Avatar, AvatarFallback, Badge } from "@/components/ui/primitives";
import { cn } from "@/lib/utils/cn";
import type { AdminNavigationContext, AdminRole, AuthUser } from "@/types/auth";

const navGroups = [
  { title: "OVERVIEW", items: [{ label: "Dashboard", href: "/admin", icon: LayoutDashboard }] },
  {
    title: "CONTENT",
    items: [
      { label: "Berita", href: "/admin/berita", icon: BookOpenText },
      { label: "Kategori Berita", href: "/admin/kategori-berita", icon: FolderOpen },
      { label: "Galeri", href: "/admin/galeri", icon: GalleryVerticalEnd },
      { label: "Pengumuman", href: "/admin/pengumuman", icon: Megaphone },
    ],
  },
  {
    title: "ORGANISASI",
    items: [
      { label: "Struktur Organisasi", href: "/admin/struktur", icon: Network },
      { label: "Divisi", href: "/admin/divisi", icon: UsersRound },
      { label: "Pengurus", href: "/admin/pengurus", icon: Users },
      { label: "Periode Kepengurusan", href: "/admin/periode", icon: CalendarDays },
    ],
  },
  {
    title: "MANAJEMEN",
    items: [
      { label: "Pendaftaran", href: "/admin/pendaftaran", icon: FileText },
      { label: "Aspirasi", href: "/admin/aspirasi", icon: MessageSquareText },
      { label: "Arsip Dokumen", href: "/admin/arsip-dokumen", icon: FolderOpen },
    ],
  },
  {
    title: "SYSTEM",
    items: [
      { label: "Pengaturan Landing Page", href: "/admin/settings", icon: Settings2 },
      { label: "Manajemen Pengguna", href: "/admin/users", icon: Users },
      { label: "Activity Log", href: "/admin/activity-logs", icon: Activity },
    ],
  },
] as const;

function visibleForRole(role: AdminRole, href: string) {
  if (role === "SUPER_ADMIN") return true;
  if (href === "/admin/users" || href === "/admin/activity-logs") return false;
  if (href === "/admin/settings") return role === "BPH";
  if (role === "BPH") return true;
  return true;
}

function SidebarNavigation({ onNavigate, role }: { onNavigate?: () => void; role: AdminRole }) {
  const pathname = usePathname();
  return (
    <nav aria-label="Navigasi admin" className="space-y-5 px-3 py-5">
      {navGroups.map((group) => (
        <section key={group.title} aria-labelledby={`nav-${group.title}`}>
          <h2
            id={`nav-${group.title}`}
            className="px-3 pb-2 text-[10px] font-extrabold tracking-[0.14em] text-muted"
          >
            {group.title}
          </h2>
          <ul className="space-y-1">
            {group.items
              .filter(({ href }) => visibleForRole(role, href))
              .map(({ label, href, icon: Icon }) => {
                const active = pathname === href || (href !== "/admin" && pathname.startsWith(`${href}/`));
                return (
                  <li key={label}>
                    <Link
                      onClick={onNavigate}
                      aria-current={active ? "page" : undefined}
                      href={href}
                      className={cn(
                        "flex min-h-10 items-center gap-3 rounded-md px-3 text-[13px] font-medium text-body transition-colors hover:bg-white/80 hover:text-primary aria-[current=page]:bg-primary-container aria-[current=page]:font-semibold aria-[current=page]:text-white",
                      )}
                    >
                      <Icon aria-hidden="true" className="size-[17px] shrink-0" />
                      <span>{label}</span>
                    </Link>
                  </li>
                );
              })}
          </ul>
        </section>
      ))}
    </nav>
  );
}

function SidebarBrand() {
  return (
    <Link
      href="/admin"
      className="flex h-[4.5rem] items-center gap-3 border-b border-border/60 px-5"
      aria-label="BEM FMIKOM Admin Portal"
    >
      <span className="grid size-10 place-items-center rounded-lg bg-primary font-heading text-xs font-extrabold text-white">
        BEM
      </span>
      <span className="flex flex-col leading-tight">
        <strong className="font-heading text-sm font-bold text-primary">BEM FMIKOM</strong>
        <span className="text-[10px] font-medium tracking-wide text-muted">ADMIN PORTAL</span>
      </span>
    </Link>
  );
}

export function AdminSidebar({ onNavigate, role }: { onNavigate?: () => void; role: AdminRole }) {
  return (
    <aside className="flex h-full w-full flex-col bg-surface-muted">
      <SidebarBrand />
      <div className="flex-1 overflow-y-auto">
        <SidebarNavigation onNavigate={onNavigate} role={role} />
      </div>
      <div className="border-t border-border/60 p-3">
        <Link
          href="/"
          className="flex min-h-10 items-center gap-3 rounded-md px-3 text-xs font-semibold text-primary hover:bg-white"
        >
          <Network aria-hidden="true" className="size-4" />
          Lihat Web Publik
        </Link>
      </div>
    </aside>
  );
}

export function AdminTopbar({
  onMenuClick,
  user,
  onLogout,
}: {
  onMenuClick: () => void;
  user: AuthUser;
  onLogout: () => void;
}) {
  const divisionName = user.division?.name.trim();
  const divisionScope = divisionName ? `Scope: ${divisionName}` : "Scope divisi tidak tersedia";

  return (
    <header className="sticky top-0 z-30 flex h-[4.5rem] items-center justify-between border-b border-border/70 bg-surface px-4 md:px-6">
      <div className="flex min-w-0 items-center gap-3">
        <Button
          aria-label="Buka navigasi admin"
          onClick={onMenuClick}
          variant="outline"
          size="icon"
          className="shrink-0 lg:hidden"
        >
          <Menu aria-hidden="true" className="size-5" />
        </Button>
        <div className="hidden size-9 place-items-center rounded-md bg-primary text-[10px] font-extrabold text-white sm:grid lg:hidden">
          BEM
        </div>
        <div className="min-w-0">
          <p className="truncate font-heading text-sm font-bold text-primary md:text-base">
            BEM FMIKOM UNUGHA
          </p>
          <p className="hidden text-xs text-muted sm:block">Fakultas Matematika dan Ilmu Komputer</p>
        </div>
      </div>
      <div className="flex min-w-0 shrink-0 items-center gap-2 sm:gap-3">
        {user.role === "KEPALA_BIDANG" ? (
          <Badge
            aria-label={divisionScope}
            className="max-w-[34vw] min-w-0 shrink border border-secondary/25 bg-warning/15 px-2 text-primary sm:max-w-56 sm:px-2.5"
            title={divisionScope}
            variant="warning"
          >
            <span className="block min-w-0 truncate">{divisionScope}</span>
          </Badge>
        ) : null}
        <span className="hidden items-center gap-1.5 rounded-full bg-surface-muted px-3 py-1.5 text-xs font-medium text-muted sm:inline-flex">
          <ShieldCheck aria-hidden="true" className="size-3.5 text-primary" />
          Preview fondasi
        </span>
        <span
          aria-hidden="true"
          className="relative hidden size-9 place-items-center rounded-md text-muted sm:grid"
        >
          <Bell className="size-[18px]" />
          <span className="absolute right-2 top-2 size-1.5 rounded-full bg-secondary" />
        </span>
        <span className="hidden h-8 w-px bg-border sm:block" />
        <div className="flex items-center gap-2">
          <span className="hidden text-right leading-tight sm:block">
            <span className="block text-xs font-semibold text-primary">{user.name}</span>
            <span className="text-[10px] text-muted">{user.role.replaceAll("_", " ")}</span>
          </span>
          <Avatar className="size-9 rounded-full">
            <AvatarFallback>{user.name.slice(0, 2).toUpperCase()}</AvatarFallback>
          </Avatar>
        </div>
        <Button aria-label="Keluar dari akun" onClick={onLogout} variant="ghost" size="icon">
          <LogOut aria-hidden="true" className="size-4" />
        </Button>
      </div>
    </header>
  );
}

export function AdminContent({ children }: { children: React.ReactNode }) {
  return (
    <main id="admin-main" className="min-h-[calc(100dvh-7rem)]">
      {children}
    </main>
  );
}

export function AdminMobileDrawer({
  open,
  onOpenChange,
  role,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  role: AdminRole;
}) {
  return (
    <Drawer
      open={open}
      onOpenChange={onOpenChange}
      title="Admin Portal"
      side="left"
      className="w-[min(86vw,18rem)] p-0"
    >
      <AdminSidebar role={role} onNavigate={() => onOpenChange(false)} />
    </Drawer>
  );
}

export function AdminShell({
  children,
  context,
  user,
  onLogout,
}: {
  children: React.ReactNode;
  context?: AdminNavigationContext;
  user: AuthUser;
  onLogout: () => void;
}) {
  const [open, setOpen] = React.useState(false);
  const role = context?.role ?? user?.role;
  if (!role) return null;
  return (
    <div className="min-h-dvh bg-background" data-admin-role={role}>
      <div className="hidden lg:fixed lg:inset-y-0 lg:left-0 lg:z-20 lg:block lg:w-[var(--sidebar-width)]">
        <AdminSidebar role={role} />
      </div>
      <div className="lg:pl-[var(--sidebar-width)]">
        <AdminTopbar onMenuClick={() => setOpen(true)} user={user} onLogout={onLogout} />
        <div className="border-b border-border/70 bg-primary-container/10 px-4 py-2.5 md:px-6">
          <div className="mx-auto flex max-w-[1440px] items-center gap-2 text-xs">
            <span className="size-2 rounded-full bg-secondary" />
            <span className="font-semibold uppercase tracking-wide text-muted">Mode akses</span>
            <span className="font-bold text-primary">Sesi terverifikasi</span>
          </div>
        </div>
        <AdminContent>{children}</AdminContent>
      </div>
      <AdminMobileDrawer open={open} onOpenChange={setOpen} role={role} />
    </div>
  );
}

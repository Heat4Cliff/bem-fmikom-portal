"use client";

import { Clock3, Search, ShieldCheck } from "lucide-react";
import * as React from "react";
import { AdminPageHeader } from "@/components/admin/admin-page-header";
import { Button } from "@/components/ui/button";
import { Badge, Input, Select } from "@/components/ui/primitives";
import { Alert } from "@/components/ui/primitives";
import { useAuth } from "@/lib/auth/auth-provider";
import { getActivities, getActivity, type Activity } from "./api";

const titleCase = (value: string) =>
  value
    .toLowerCase()
    .replaceAll("_", " ")
    .replace(/\b\w/g, (c) => c.toUpperCase());
const stamp = (value: string) =>
  new Intl.DateTimeFormat("id-ID", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "Asia/Jakarta",
  }).format(new Date(value));
function detailEntries(value: unknown, path = ""): Array<[string, string]> {
  if (value === null || value === undefined) return [];
  if (Array.isArray(value)) return value.flatMap((x, i) => detailEntries(x, `${path}[${i + 1}]`));
  if (typeof value === "object")
    return Object.entries(value as Record<string, unknown>).flatMap(([k, v]) =>
      detailEntries(v, path ? `${path} · ${titleCase(k)}` : titleCase(k)),
    );
  return [[path || "Detail", typeof value === "string" ? value : String(value)]];
}
const actionTone = (action: string) =>
  action === "DELETE" || action === "DEACTIVATE"
    ? "danger"
    : action === "UPLOAD" || action === "PUBLISH"
      ? "warning"
      : action === "LOGIN" || action === "CREATE"
        ? "info"
        : "neutral";

export function ActivityAdmin() {
  const { user } = useAuth();
  const allowed = user?.role === "SUPER_ADMIN";
  const [items, setItems] = React.useState<Activity[]>([]);
  const [page, setPage] = React.useState(1);
  const [pages, setPages] = React.useState(1);
  const [search, setSearch] = React.useState("");
  const [action, setAction] = React.useState("");
  const [entity, setEntity] = React.useState("");
  const [dateFrom, setDateFrom] = React.useState("");
  const [dateTo, setDateTo] = React.useState("");
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState("");
  const [selected, setSelected] = React.useState<Activity | null>(null);
  const [detailLoading, setDetailLoading] = React.useState(false);
  const load = React.useCallback(async () => {
    if (!allowed) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError("");
    try {
      const r = await getActivities({
        page,
        limit: 20,
        search,
        action,
        entity_type: entity,
        date_from: dateFrom,
        date_to: dateTo,
      });
      setItems(r.items);
      setPages(Math.max(1, r.meta.total_pages));
    } catch {
      setError("Activity log gagal dimuat atau akses akun tidak tersedia.");
    } finally {
      setLoading(false);
    }
  }, [allowed, page, search, action, entity, dateFrom, dateTo]);
  React.useEffect(() => {
    void load();
  }, [load]);
  const openDetail = async (id: number) => {
    setDetailLoading(true);
    try {
      setSelected(await getActivity(id));
    } catch {
      setError("Detail activity log tidak dapat dibuka.");
    } finally {
      setDetailLoading(false);
    }
  };
  const detailButton = (item: Activity) => (
    <Button size="sm" variant="outline" onClick={() => void openDetail(item.id)}>
      Detail audit →
    </Button>
  );
  const description = (a: Activity) =>
    detailEntries(a.details)
      .slice(0, 2)
      .map(([, v]) => v)
      .join(" · ") || `${titleCase(a.entity_type)} ${a.entity_id ?? ""}`;
  if (!allowed)
    return (
      <div className="p-6">
        <Alert variant="danger" title="Akses ditolak">
          Activity Log hanya tersedia untuk SUPER_ADMIN.
        </Alert>
      </div>
    );
  return (
    <div className="space-y-5 p-4 md:p-6 lg:p-8">
      <AdminPageHeader
        title="Activity Log / Audit Trail"
        description="Riwayat operasi administrator. Data hanya dapat dibaca dan tidak memiliki kontrol ubah atau hapus."
      />
      <div className="flex items-center gap-3 rounded-xl bg-primary p-4 text-white shadow-card">
        <ShieldCheck className="size-7 shrink-0 text-secondary-container" />
        <div>
          <p className="text-sm font-bold">Audit event ledger · akses baca saja</p>
          <p className="mt-1 text-xs text-blue-100">
            Metadata sensitif disensor pada respons. IP dan User-Agent tidak ditampilkan.
          </p>
        </div>
      </div>
      {error && (
        <p role="alert" className="rounded-md bg-danger/10 p-3 text-sm text-danger">
          {error}
        </p>
      )}
      <section
        aria-label="Filter activity log"
        className="grid gap-2 rounded-xl border border-border bg-surface p-3 sm:grid-cols-2 xl:grid-cols-[minmax(14rem,1fr)_repeat(4,minmax(8rem,.5fr))]"
      >
        <div className="relative">
          <span className="sr-only">Cari actor, entitas atau ID</span>
          <Search className="absolute left-3 top-3 size-4 text-muted" />
          <Input
            aria-label="Cari actor, entitas atau ID"
            className="pl-9"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
            placeholder="Cari actor, entitas, detail…"
          />
        </div>
        <div>
          <span className="sr-only">Filter aksi</span>
          <Select
            aria-label="Filter aksi"
            value={action}
            onChange={(e) => {
              setAction(e.target.value);
              setPage(1);
            }}
          >
            <option value="">Semua aksi</option>
            {[
              "LOGIN",
              "LOGOUT",
              "CREATE",
              "UPDATE",
              "DELETE",
              "PUBLISH",
              "UNPUBLISH",
              "ACTIVATE",
              "DEACTIVATE",
              "UPLOAD",
              "DOWNLOAD",
              "EXPORT",
            ].map((x) => (
              <option key={x}>{x}</option>
            ))}
          </Select>
        </div>
        <div>
          <span className="sr-only">Filter tipe entitas</span>
          <Input
            aria-label="Tipe entitas"
            value={entity}
            onChange={(e) => {
              setEntity(e.target.value.toUpperCase());
              setPage(1);
            }}
            placeholder="Tipe entitas"
          />
        </div>
        <div className="grid gap-1 text-xs text-muted">
          Dari
          <Input
            aria-label="Dari tanggal"
            type="date"
            value={dateFrom}
            onChange={(e) => {
              setDateFrom(e.target.value);
              setPage(1);
            }}
          />
        </div>
        <div className="grid gap-1 text-xs text-muted">
          Sampai
          <Input
            aria-label="Sampai tanggal"
            type="date"
            value={dateTo}
            onChange={(e) => {
              setDateTo(e.target.value);
              setPage(1);
            }}
          />
        </div>
        <Button
          className="sm:col-span-2 xl:col-span-5"
          variant="outline"
          onClick={() => {
            setSearch("");
            setAction("");
            setEntity("");
            setDateFrom("");
            setDateTo("");
            setPage(1);
          }}
        >
          Reset filter
        </Button>
      </section>
      <section
        aria-label="Riwayat aktivitas"
        className="overflow-hidden rounded-xl border border-border bg-surface"
      >
        {loading ? (
          <p role="status" className="p-6 text-sm text-muted">
            Memuat activity log…
          </p>
        ) : items.length === 0 ? (
          <p className="p-8 text-center text-sm text-muted">Belum ada activity log untuk filter ini.</p>
        ) : (
          <>
            <div className="hidden overflow-x-auto md:block">
              <table className="w-full min-w-[850px] text-left text-sm">
                <thead className="bg-surface-muted text-xs uppercase tracking-wide text-muted">
                  <tr>
                    {["Waktu", "Actor", "Aksi", "Entitas", "Ringkasan", "Detail"].map((x) => (
                      <th key={x} className="px-4 py-3">
                        {x}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {items.map((item) => (
                    <tr key={item.id} className="hover:bg-surface-muted/50">
                      <td className="whitespace-nowrap px-4 py-3 text-xs">{stamp(item.timestamp)}</td>
                      <td className="px-4 py-3">
                        <p className="font-semibold text-primary">
                          {item.actor?.name ?? "Akun tidak tersedia"}
                        </p>
                        <p className="text-xs text-muted">
                          {item.actor ? `${item.actor.role} · ${item.actor.email}` : "Actor dihapus"}
                        </p>
                      </td>
                      <td className="px-4 py-3">
                        <Badge variant={actionTone(item.action)}>{item.action}</Badge>
                      </td>
                      <td className="px-4 py-3">
                        {titleCase(item.entity_type)}
                        <p className="text-xs text-muted">{item.entity_id ?? "—"}</p>
                      </td>
                      <td className="max-w-72 px-4 py-3">
                        <span className="line-clamp-2 text-body">{description(item)}</span>
                      </td>
                      <td className="px-4 py-3">{detailButton(item)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <div className="grid gap-3 p-3 md:hidden">
              {items.map((item) => (
                <article key={item.id} className="rounded-xl border border-border p-4">
                  <div className="flex items-center justify-between gap-2">
                    <Badge variant={actionTone(item.action)}>{item.action}</Badge>
                    <span className="inline-flex items-center gap-1 text-xs text-muted">
                      <Clock3 className="size-3.5" />
                      {stamp(item.timestamp)}
                    </span>
                  </div>
                  <h2 className="mt-3 font-heading text-lg font-semibold text-primary">
                    {titleCase(item.entity_type)}
                    {item.entity_id ? ` · ${item.entity_id}` : ""}
                  </h2>
                  <p className="mt-1 text-sm">
                    {item.actor?.name ?? "Akun tidak tersedia"}{" "}
                    <span className="text-muted">{item.actor ? `(${item.actor.role})` : ""}</span>
                  </p>
                  <p className="mt-3 rounded-md bg-surface-muted p-3 text-sm leading-6 text-body">
                    {description(item)}
                  </p>
                  <div className="mt-3 flex justify-end">{detailButton(item)}</div>
                </article>
              ))}
            </div>
            <div className="flex items-center justify-between border-t border-border bg-surface-muted px-4 py-3 text-sm">
              <span>
                Halaman {page} dari {pages}
              </span>
              <div className="flex gap-2">
                <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage(page - 1)}>
                  Sebelumnya
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  disabled={page >= pages}
                  onClick={() => setPage(page + 1)}
                >
                  Berikutnya
                </Button>
              </div>
            </div>
          </>
        )}
      </section>
      <div className="sr-only" aria-live="polite">
        {detailLoading ? "Memuat detail audit" : ""}
      </div>
      {selected && (
        <div className="fixed inset-0 z-50 flex justify-end bg-slate-950/45">
          <button
            type="button"
            aria-label="Tutup detail audit"
            className="absolute inset-0 size-full cursor-default"
            onClick={() => setSelected(null)}
          />
          <section
            role="dialog"
            aria-modal="true"
            aria-labelledby="activity-detail-title"
            className="relative z-10 flex h-full w-full max-w-2xl flex-col overflow-y-auto border-l border-border bg-surface shadow-float"
          >
            <header className="sticky top-0 flex items-start justify-between gap-4 border-b border-border bg-surface p-5">
              <div>
                <Badge variant={actionTone(selected.action)}>{selected.action}</Badge>
                <h2 id="activity-detail-title" className="mt-3 font-heading text-xl font-bold text-primary">
                  Detail Log Aktivitas #{selected.id}
                </h2>
                <p className="mt-1 text-sm text-muted">{stamp(selected.timestamp)}</p>
              </div>
              <Button variant="outline" onClick={() => setSelected(null)}>
                Tutup
              </Button>
            </header>
            <div className="flex-1 space-y-5 p-5">
              <section className="rounded-lg bg-surface-muted p-4">
                <h3 className="text-xs font-bold uppercase tracking-wide text-muted">Identitas actor</h3>
                {selected.actor ? (
                  <dl className="mt-3 grid gap-3 sm:grid-cols-2">
                    <div>
                      <dt className="text-xs text-muted">Nama</dt>
                      <dd className="font-semibold text-primary">{selected.actor.name}</dd>
                    </div>
                    <div>
                      <dt className="text-xs text-muted">Role</dt>
                      <dd>{selected.actor.role}</dd>
                    </div>
                    <div className="sm:col-span-2">
                      <dt className="text-xs text-muted">Email</dt>
                      <dd className="break-all">{selected.actor.email}</dd>
                    </div>
                  </dl>
                ) : (
                  <p className="mt-2 text-sm">Akun actor tidak tersedia.</p>
                )}
              </section>
              <section>
                <h3 className="text-xs font-bold uppercase tracking-wide text-muted">Entitas & target</h3>
                <dl className="mt-3 grid gap-3 rounded-lg border border-border p-4 sm:grid-cols-2">
                  <div>
                    <dt className="text-xs text-muted">Tipe</dt>
                    <dd>{titleCase(selected.entity_type)}</dd>
                  </div>
                  <div>
                    <dt className="text-xs text-muted">ID</dt>
                    <dd>{selected.entity_id ?? "—"}</dd>
                  </div>
                  <div className="sm:col-span-2">
                    <dt className="text-xs text-muted">Waktu (UTC)</dt>
                    <dd>{new Date(selected.timestamp).toISOString()}</dd>
                  </div>
                </dl>
              </section>
              <section>
                <h3 className="text-xs font-bold uppercase tracking-wide text-muted">Detail perubahan</h3>
                {detailEntries(selected.details).length ? (
                  <dl className="mt-3 grid gap-2">
                    {detailEntries(selected.details).map(([key, value]) => (
                      <div
                        key={`${key}-${value}`}
                        className="grid gap-1 rounded-md bg-surface-muted p-3 sm:grid-cols-[minmax(8rem,.45fr)_1fr]"
                      >
                        <dt className="text-xs font-semibold text-muted">{key}</dt>
                        <dd className="break-words text-sm text-body">{value}</dd>
                      </div>
                    ))}
                  </dl>
                ) : (
                  <p className="mt-3 rounded-md border border-dashed border-border p-4 text-sm text-muted">
                    Detail tambahan tidak tersedia.
                  </p>
                )}
              </section>
              <p className="rounded-lg border border-primary/15 bg-primary/5 p-4 text-sm leading-6 text-body">
                Log ditampilkan sebagai data baca saja. Metadata IP/User-Agent tidak disajikan dan nilai
                detail sensitif telah disensor oleh API.
              </p>
            </div>
            <footer className="sticky bottom-0 border-t border-border bg-surface p-4 text-xs text-muted">
              Tidak tersedia tindakan edit atau hapus.
            </footer>
          </section>
        </div>
      )}
    </div>
  );
}

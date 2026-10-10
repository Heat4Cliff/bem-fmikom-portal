"use client";

import { CalendarDays, ChevronLeft, ChevronRight, Search, ShieldCheck } from "lucide-react";
import Link from "next/link";
import * as React from "react";
import { Badge, Input, Label, Select } from "@/components/ui/primitives";
import { Button } from "@/components/ui/button";
import { getDivisions } from "@/features/organization/api";
import type { Division } from "@/features/organization/types";
import { getAspiration, getAspirations, updateAspirationStatus } from "./api";
import type { Aspiration, AspirationStatus } from "./types";
import { aspirationStatusLabel } from "./types";

const pageSize = 20;

export function AspirationAdminList() {
  const [items, setItems] = React.useState<Aspiration[]>([]);
  const [divisions, setDivisions] = React.useState<Division[]>([]);
  const [search, setSearch] = React.useState("");
  const [status, setStatus] = React.useState("");
  const [divisionID, setDivisionID] = React.useState("");
  const [dateFrom, setDateFrom] = React.useState("");
  const [dateTo, setDateTo] = React.useState("");
  const [page, setPage] = React.useState(1);
  const [total, setTotal] = React.useState(0);
  const [totalPages, setTotalPages] = React.useState(0);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState("");

  React.useEffect(() => {
    let active = true;
    getDivisions({ page: 1, limit: 100 })
      .then((result) => {
        if (active) setDivisions(result.items);
      })
      .catch(() => {
        if (active) setDivisions([]);
      });
    return () => {
      active = false;
    };
  }, []);

  React.useEffect(() => {
    let active = true;
    setLoading(true);
    setError("");
    getAspirations({
      page,
      limit: pageSize,
      search,
      status,
      division_id: divisionID,
      date_from: dateFrom,
      date_to: dateTo,
    })
      .then((result) => {
        if (!active) return;
        setItems(result.items);
        setTotal(result.meta.total);
        setTotalPages(result.meta.total_pages);
      })
      .catch(() => {
        if (active) setError("Data aspirasi gagal dimuat. Periksa koneksi atau akses akun.");
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [dateFrom, dateTo, divisionID, page, search, status]);

  function changeFilter(setter: (value: string) => void, value: string) {
    setter(value);
    setPage(1);
  }

  return (
    <div className="mx-auto w-full max-w-[1600px] p-4 md:p-6 lg:p-8">
      <nav aria-label="Breadcrumb" className="mb-4 text-xs text-muted">
        <Link className="hover:text-primary" href="/admin">
          Dashboard
        </Link>
        <span className="mx-2">/</span>
        <span className="text-primary">Aspirasi</span>
      </nav>
      <header className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-xs font-bold uppercase tracking-[0.14em] text-secondary">Manajemen</p>
          <h1 className="mt-1 font-heading text-2xl font-bold text-primary sm:text-3xl">
            Manajemen Aspirasi
          </h1>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-muted">
            Tinjau masukan mahasiswa yang diteruskan ke divisi tujuan.
          </p>
        </div>
        <div className="inline-flex items-center gap-2 rounded-lg bg-primary/5 px-3 py-2 text-xs text-primary">
          <ShieldCheck aria-hidden="true" className="size-4" />
          Identitas pengirim anonim tidak disimpan
        </div>
      </header>

      <section
        aria-label="Filter aspirasi"
        className="mb-4 grid gap-3 rounded-xl border border-border bg-white p-4 shadow-card sm:grid-cols-2 xl:grid-cols-[minmax(14rem,1.5fr)_repeat(4,minmax(9rem,1fr))]"
      >
        <div className="relative sm:col-span-2 xl:col-span-1">
          <Label htmlFor="aspiration-search" className="sr-only">
            Cari aspirasi
          </Label>
          <Search aria-hidden="true" className="absolute left-3 top-3 size-4 text-muted" />
          <Input
            id="aspiration-search"
            value={search}
            onChange={(event) => changeFilter(setSearch, event.target.value)}
            placeholder="Cari judul atau pengirim…"
            className="pl-9"
          />
        </div>
        <div>
          <Label htmlFor="aspiration-filter-division" className="sr-only">
            Tujuan divisi
          </Label>
          <Select
            id="aspiration-filter-division"
            value={divisionID}
            onChange={(event) => changeFilter(setDivisionID, event.target.value)}
          >
            <option value="">Semua divisi</option>
            {divisions.map((division) => (
              <option key={division.id} value={division.id}>
                {division.name}
              </option>
            ))}
          </Select>
        </div>
        <div>
          <Label htmlFor="aspiration-filter-status" className="sr-only">
            Status
          </Label>
          <Select
            id="aspiration-filter-status"
            value={status}
            onChange={(event) => changeFilter(setStatus, event.target.value)}
          >
            <option value="">Semua status</option>
            <option value="PENDING">Menunggu</option>
            <option value="READ">Dibaca</option>
            <option value="RESOLVED">Selesai</option>
          </Select>
        </div>
        <div>
          <Label htmlFor="aspiration-filter-from" className="sr-only">
            Tanggal masuk dari
          </Label>
          <Input
            id="aspiration-filter-from"
            type="date"
            value={dateFrom}
            onChange={(event) => changeFilter(setDateFrom, event.target.value)}
            aria-label="Tanggal masuk dari"
          />
        </div>
        <div>
          <Label htmlFor="aspiration-filter-to" className="sr-only">
            Tanggal masuk sampai
          </Label>
          <Input
            id="aspiration-filter-to"
            type="date"
            value={dateTo}
            onChange={(event) => changeFilter(setDateTo, event.target.value)}
            aria-label="Tanggal masuk sampai"
          />
        </div>
      </section>

      <div className="mb-3 flex items-center justify-between text-sm text-muted">
        <span>{loading ? "Memuat aspirasi…" : `${total} aspirasi`}</span>
        <span className="hidden sm:inline">Pilih tiket untuk melihat detail</span>
      </div>
      {error ? (
        <p
          role="alert"
          className="mb-4 rounded-md border border-danger/25 bg-danger/5 p-3 text-sm text-red-900"
        >
          {error}
        </p>
      ) : null}
      {!loading && !error && items.length === 0 ? (
        <div className="rounded-xl border border-dashed border-border bg-white px-6 py-14 text-center">
          <p className="font-semibold text-primary">Belum ada aspirasi yang cocok</p>
          <p className="mt-1 text-sm text-muted">Ubah filter pencarian atau periksa kembali nanti.</p>
        </div>
      ) : null}

      <div className="hidden overflow-hidden rounded-xl border border-border bg-white shadow-card md:block">
        <div className="overflow-x-auto">
          <table className="w-full min-w-[850px] border-collapse text-left text-sm">
            <thead className="bg-surface-muted text-[11px] uppercase tracking-wide text-muted">
              <tr>
                <th scope="col" className="px-5 py-3">
                  Judul
                </th>
                <th scope="col" className="px-4 py-3">
                  Divisi tujuan
                </th>
                <th scope="col" className="px-4 py-3">
                  Pengirim
                </th>
                <th scope="col" className="px-4 py-3">
                  Status
                </th>
                <th scope="col" className="px-4 py-3">
                  Tanggal masuk
                </th>
                <th scope="col" className="px-4 py-3">
                  <span className="sr-only">Aksi</span>
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {items.map((item) => (
                <tr key={item.id} className="align-top hover:bg-surface/70">
                  <td className="max-w-[28rem] px-5 py-4">
                    <Link
                      href={`/admin/aspirasi/${item.id}`}
                      className="font-semibold text-primary hover:underline"
                    >
                      {item.subject}
                    </Link>
                    <p className="mt-1 line-clamp-2 text-xs leading-5 text-muted">{item.message}</p>
                    <p className="mt-1 text-[11px] text-muted">ASP-{item.id}</p>
                  </td>
                  <td className="px-4 py-4 text-body">{item.division_name}</td>
                  <td className="px-4 py-4">
                    {senderLabel(item)}
                    {item.nim ? <p className="mt-1 text-xs text-muted">NIM: {item.nim}</p> : null}
                  </td>
                  <td className="px-4 py-4">
                    <StatusBadge status={item.status} />
                  </td>
                  <td className="px-4 py-4 whitespace-nowrap text-xs text-body">
                    {formatDate(item.submitted_at)}
                  </td>
                  <td className="px-4 py-4">
                    <Button asChild variant="outline" size="sm">
                      <Link href={`/admin/aspirasi/${item.id}`}>Detail</Link>
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <div className="space-y-3 md:hidden">
        {items.map((item) => (
          <article key={item.id} className="rounded-xl border border-border bg-white p-4 shadow-card">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <StatusBadge status={item.status} />
              <time className="text-xs text-muted">{formatDate(item.submitted_at)}</time>
            </div>
            <p className="mt-3 text-xs font-semibold tracking-wide text-primary">ASP-{item.id}</p>
            <h2 className="mt-1 font-heading text-lg font-semibold leading-6 text-primary">{item.subject}</h2>
            <p className="mt-2 line-clamp-3 text-sm leading-6 text-body">{item.message}</p>
            <dl className="mt-4 space-y-2 text-xs">
              <div>
                <dt className="inline text-muted">Tujuan: </dt>
                <dd className="inline font-medium text-primary">{item.division_name}</dd>
              </div>
              <div>
                <dt className="inline text-muted">Pengirim: </dt>
                <dd className="inline text-body">
                  {senderLabel(item)}
                  {item.nim ? ` (NIM: ${item.nim})` : ""}
                </dd>
              </div>
            </dl>
            <Button asChild className="mt-4 min-h-11 w-full">
              <Link href={`/admin/aspirasi/${item.id}`}>Lihat detail</Link>
            </Button>
          </article>
        ))}
      </div>

      <nav
        aria-label="Halaman aspirasi"
        className="mt-5 flex items-center justify-between gap-3 rounded-xl bg-white p-3 shadow-card"
      >
        <span className="text-xs text-muted">
          Halaman {page}
          {totalPages ? ` dari ${totalPages}` : ""}
        </span>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1 || loading}
            onClick={() => setPage((value) => Math.max(1, value - 1))}
          >
            <ChevronLeft aria-hidden="true" className="size-4" />
            <span className="hidden sm:inline">Sebelumnya</span>
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={!totalPages || page >= totalPages || loading}
            onClick={() => setPage((value) => value + 1)}
          >
            <span className="hidden sm:inline">Selanjutnya</span>
            <ChevronRight aria-hidden="true" className="size-4" />
          </Button>
        </div>
      </nav>
    </div>
  );
}

export function StatusBadge({ status }: { status: AspirationStatus }) {
  const variant = status === "PENDING" ? "warning" : status === "READ" ? "info" : "neutral";
  const color = status === "RESOLVED" ? "border-slate-300 bg-slate-100 text-slate-700" : undefined;
  return (
    <Badge variant={variant} className={color}>
      {aspirationStatusLabel[status]}
    </Badge>
  );
}

export function AspirationDetail({ id }: { id: number }) {
  const [item, setItem] = React.useState<Aspiration | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState("");
  const [saving, setSaving] = React.useState(false);
  const [notice, setNotice] = React.useState("");

  const refresh = React.useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      setItem(await getAspiration(id));
    } catch {
      setError("Detail aspirasi tidak dapat dimuat. Resource mungkin tidak tersedia untuk akun ini.");
    } finally {
      setLoading(false);
    }
  }, [id]);
  React.useEffect(() => {
    void refresh();
  }, [refresh]);

  async function updateStatus(status: "READ" | "RESOLVED") {
    setSaving(true);
    setNotice("");
    setError("");
    try {
      setItem(await updateAspirationStatus(id, status));
      setNotice(`Status aspirasi diperbarui menjadi ${aspirationStatusLabel[status]}.`);
    } catch {
      setError("Status aspirasi gagal diperbarui. Periksa kembali akses atau status terkini.");
    } finally {
      setSaving(false);
    }
  }

  if (loading)
    return (
      <div className="mx-auto max-w-5xl p-4 md:p-8" role="status">
        Memuat detail aspirasi…
      </div>
    );
  if (error && !item)
    return (
      <div className="mx-auto max-w-5xl p-4 md:p-8">
        <p role="alert" className="rounded-md bg-danger/5 p-4 text-sm text-red-900">
          {error}
        </p>
        <Button asChild variant="outline" className="mt-4">
          <Link href="/admin/aspirasi">Kembali ke daftar</Link>
        </Button>
      </div>
    );
  if (!item) return null;

  return (
    <div className="mx-auto max-w-6xl p-4 md:p-6 lg:p-8">
      <nav
        aria-label="Breadcrumb"
        className="mb-4 flex flex-wrap items-center justify-between gap-3 text-xs text-muted"
      >
        <span>
          <Link href="/admin">Dashboard</Link>
          <span className="mx-2">/</span>
          <Link href="/admin/aspirasi">Manajemen Aspirasi</Link>
          <span className="mx-2">/</span>
          <span className="text-primary">ASP-{item.id}</span>
        </span>
        <Button asChild variant="outline" size="sm">
          <Link href="/admin/aspirasi">Kembali ke daftar</Link>
        </Button>
      </nav>
      <header className="mb-5 rounded-2xl border border-border bg-gradient-to-br from-white to-primary/5 p-5 sm:p-7">
        <div className="flex flex-wrap items-center gap-2">
          <StatusBadge status={item.status} />
          <Badge variant="neutral">{item.division_name}</Badge>
        </div>
        <p className="mt-4 text-xs font-semibold uppercase tracking-wide text-muted">Tiket ASP-{item.id}</p>
        <h1 className="mt-1 max-w-4xl break-words font-heading text-2xl font-bold text-primary sm:text-3xl">
          {item.subject}
        </h1>
      </header>

      {notice ? (
        <p
          role="status"
          className="mb-4 rounded-md border border-success/25 bg-success/5 p-3 text-sm text-emerald-900"
        >
          {notice}
        </p>
      ) : null}
      {error ? (
        <p
          role="alert"
          className="mb-4 rounded-md border border-danger/25 bg-danger/5 p-3 text-sm text-red-900"
        >
          {error}
        </p>
      ) : null}
      <div className="grid items-start gap-5 lg:grid-cols-[minmax(0,1fr)_18rem]">
        <div className="space-y-5">
          <section className="rounded-xl border border-border bg-white p-5 shadow-card sm:p-7">
            <h2 className="font-heading text-lg font-bold text-primary">Uraian aspirasi</h2>
            <p className="mt-4 whitespace-pre-wrap break-words text-sm leading-7 text-body">{item.message}</p>
          </section>
          <section className="rounded-xl border border-border bg-white p-5 shadow-card sm:p-7">
            <h2 className="font-heading text-lg font-bold text-primary">Linimasa status</h2>
            <ol className="mt-5 space-y-5 border-l-2 border-primary/15 pl-5">
              <TimelineItem title="Aspirasi dikirim" at={item.submitted_at} complete />
              <TimelineItem title="Dibaca" at={item.read_at} complete={Boolean(item.read_at)} />
              <TimelineItem title="Selesai" at={item.resolved_at} complete={Boolean(item.resolved_at)} />
            </ol>
          </section>
        </div>

        <aside className="space-y-4 lg:sticky lg:top-24">
          <section className="rounded-xl border border-border bg-white p-5 shadow-card">
            <h2 className="font-heading text-lg font-bold text-primary">Informasi pengirim</h2>
            {item.is_anonymous ? (
              <p className="mt-3 rounded-lg bg-primary/5 p-3 text-sm font-medium text-primary">
                Dikirim secara anonim
              </p>
            ) : (
              <dl className="mt-3 space-y-3 text-sm">
                <DetailField label="Nama" value={item.name} />
                <DetailField label="NIM" value={item.nim} />
                <DetailField label="Email" value={item.email} />
              </dl>
            )}
            <dl className="mt-4 space-y-3 border-t border-border pt-4 text-sm">
              <DetailField label="Divisi tujuan" value={item.division_name} />
              <div>
                <dt className="text-xs text-muted">Tanggal masuk</dt>
                <dd className="mt-1 flex items-center gap-2 text-body">
                  <CalendarDays aria-hidden="true" className="size-4 text-primary" />
                  {formatDate(item.submitted_at)}
                </dd>
              </div>
            </dl>
          </section>
          {item.status !== "RESOLVED" ? (
            <section className="rounded-xl border border-border bg-white p-5 shadow-card">
              <h2 className="font-heading text-lg font-bold text-primary">Tindakan status</h2>
              <p className="mt-2 text-xs leading-5 text-muted">
                Status mengikuti alur Menunggu → Dibaca → Selesai.
              </p>
              {item.status === "PENDING" ? (
                <Button className="mt-4 w-full" disabled={saving} onClick={() => void updateStatus("READ")}>
                  {saving ? "Menyimpan…" : "Tandai Dibaca"}
                </Button>
              ) : (
                <Button
                  className="mt-4 w-full"
                  disabled={saving}
                  onClick={() => void updateStatus("RESOLVED")}
                >
                  {saving ? "Menyimpan…" : "Tandai Selesai"}
                </Button>
              )}
            </section>
          ) : null}
        </aside>
      </div>
    </div>
  );
}

function TimelineItem({ title, at, complete }: { title: string; at: string | null; complete: boolean }) {
  return (
    <li className="relative">
      <span
        aria-hidden="true"
        className={`absolute -left-[1.72rem] top-1 size-3 rounded-full border-2 ${complete ? "border-primary bg-primary" : "border-border bg-white"}`}
      />
      <p className={`text-sm font-semibold ${complete ? "text-primary" : "text-muted"}`}>{title}</p>
      <p className="mt-1 text-xs text-muted">{at ? formatDate(at) : "Menunggu"}</p>
    </li>
  );
}

function DetailField({ label, value }: { label: string; value: string | null }) {
  if (!value) return null;
  return (
    <div>
      <dt className="text-xs text-muted">{label}</dt>
      <dd className="mt-1 break-all font-medium text-body">{value}</dd>
    </div>
  );
}

function senderLabel(item: Aspiration) {
  return item.is_anonymous ? "Anonim" : item.name || item.nim || item.email || "Identitas tidak dicantumkan";
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("id-ID", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "Asia/Jakarta",
  }).format(new Date(value));
}

"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { CalendarDays, Network, Plus, Search, UsersRound } from "lucide-react";
import * as React from "react";
import { AdminBreadcrumb, AdminPageHeader } from "@/components/admin/admin-page-header";
import { MediaUpload } from "@/components/admin/media-upload";
import { ActionGroup } from "@/components/ui/primitives";
import { MediaImage } from "@/components/ui/media-image";
import { Button } from "@/components/ui/button";
import { Alert, Badge, Card, Checkbox, Input, Label, Select, Textarea } from "@/components/ui/primitives";
import { useAuth } from "@/lib/auth/auth-provider";
import {
  deleteDivision,
  deleteMember,
  getDivisions,
  getMember,
  getMembers,
  getPeriods,
  saveDivision,
  saveMember,
  savePeriod,
} from "./api";
import {
  POSITION_OPTIONS,
  positionLabel,
  type Division,
  type MemberPosition,
  type OrganizationMember,
  type Period,
} from "./types";

const pageSize = 20;
const fieldClass = "grid gap-1.5";

function Notice({ text, danger = false }: { text: string; danger?: boolean }) {
  if (!text) return null;
  return (
    <Alert className="mb-4" variant={danger ? "danger" : "success"}>
      {text}
    </Alert>
  );
}

function Confirm({
  title,
  description,
  confirmLabel,
  onCancel,
  onConfirm,
}: {
  title: string;
  description: string;
  confirmLabel: string;
  onCancel: () => void;
  onConfirm: () => void;
}) {
  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-slate-950/40 p-4" role="presentation">
      <section
        role="dialog"
        aria-modal="true"
        aria-labelledby="organization-confirm-title"
        className="w-full max-w-md rounded-xl border border-border bg-surface p-6 shadow-floating"
      >
        <h2 id="organization-confirm-title" className="font-heading text-lg font-bold text-primary">
          {title}
        </h2>
        <p className="mt-2 text-sm leading-6 text-body">{description}</p>
        <div className="mt-6 flex justify-end gap-2">
          <Button variant="outline" onClick={onCancel}>
            Batal
          </Button>
          <Button onClick={onConfirm}>{confirmLabel}</Button>
        </div>
      </section>
    </div>
  );
}

function TableFrame({ children }: { children: React.ReactNode }) {
  return <div className="overflow-x-auto rounded-lg border border-border/70 bg-surface">{children}</div>;
}

function Status({ value }: { value: string }) {
  return (
    <Badge variant={value === "ACTIVE" ? "success" : "neutral"}>
      {value === "ACTIVE" ? "Aktif" : value === "ARCHIVED" ? "Diarsipkan" : "Nonaktif"}
    </Badge>
  );
}

function PeriodEditor({
  initial,
  onCancel,
  onSaved,
}: {
  initial?: Period;
  onCancel: () => void;
  onSaved: () => void;
}) {
  const [name, setName] = React.useState(initial?.name ?? "");
  const [startDate, setStartDate] = React.useState(initial?.start_date ?? "");
  const [endDate, setEndDate] = React.useState(initial?.end_date ?? "");
  const [error, setError] = React.useState("");
  const [saving, setSaving] = React.useState(false);
  async function submit(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    setSaving(true);
    try {
      await savePeriod({ name, start_date: startDate, end_date: endDate }, initial?.id);
      onSaved();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Periode gagal disimpan.");
    } finally {
      setSaving(false);
    }
  }
  return (
    <form
      onSubmit={submit}
      className="mb-5 grid gap-4 rounded-lg border border-border bg-surface p-4 md:grid-cols-4 md:items-end md:p-5"
    >
      <div className={fieldClass}>
        <Label htmlFor="period-name">Nama Periode</Label>
        <Input
          id="period-name"
          required
          maxLength={120}
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
      </div>
      <div className={fieldClass}>
        <Label htmlFor="period-start">Tanggal Mulai</Label>
        <Input
          id="period-start"
          required
          type="date"
          value={startDate}
          onChange={(e) => setStartDate(e.target.value)}
        />
      </div>
      <div className={fieldClass}>
        <Label htmlFor="period-end">Tanggal Selesai</Label>
        <Input
          id="period-end"
          required
          type="date"
          min={startDate || undefined}
          value={endDate}
          onChange={(e) => setEndDate(e.target.value)}
        />
      </div>
      <div className="flex gap-2">
        <Button type="submit" disabled={saving}>
          {saving ? "Menyimpan…" : initial ? "Simpan Perubahan" : "Tambah Periode"}
        </Button>
        <Button type="button" variant="outline" onClick={onCancel}>
          Batal
        </Button>
      </div>
      {error ? (
        <p role="alert" className="text-sm text-danger md:col-span-4">
          {error}
        </p>
      ) : null}
    </form>
  );
}

export function PeriodManager() {
  const { user } = useAuth();
  const [items, setItems] = React.useState<Period[]>([]);
  const [search, setSearch] = React.useState("");
  const [status, setStatus] = React.useState("");
  const [editing, setEditing] = React.useState<Period | null | undefined>(undefined);
  const [notice, setNotice] = React.useState("");
  const [error, setError] = React.useState(false);
  const [pending, setPending] = React.useState<{ period: Period; action: "activate" | "archive" } | null>(
    null,
  );
  const [page, setPage] = React.useState(1);
  const [totalPages, setTotalPages] = React.useState(1);
  const canManage = user?.role === "SUPER_ADMIN";
  const refresh = React.useCallback(async () => {
    try {
      const data = await getPeriods({ page, limit: pageSize, search, status });
      setItems(data.items);
      setTotalPages(Math.max(1, data.meta.total_pages));
      setError(false);
    } catch {
      setError(true);
    }
  }, [page, search, status]);
  React.useEffect(() => {
    void refresh();
  }, [refresh]);
  async function perform() {
    if (!pending) return;
    try {
      const { activatePeriod, archivePeriod } = await import("./api");
      if (pending.action === "activate") await activatePeriod(pending.period.id);
      else await archivePeriod(pending.period.id);
      setNotice(
        pending.action === "activate" ? "Periode aktif telah diperbarui." : "Periode berhasil diarsipkan.",
      );
      setError(false);
      setPending(null);
      await refresh();
    } catch (e) {
      setNotice(e instanceof Error ? e.message : "Status periode gagal diubah.");
      setError(true);
      setPending(null);
    }
  }
  return (
    <div className="mx-auto max-w-[1440px] p-4 md:p-6 lg:p-8">
      <AdminBreadcrumb items={[{ label: "Organisasi" }, { label: "Periode Kepengurusan" }]} />
      <AdminPageHeader
        title="Periode Kepengurusan"
        description="Kelola periode organisasi dan pertahankan histori kepengurusan."
        actions={
          canManage ? (
            <Button onClick={() => setEditing(null)}>
              <Plus aria-hidden="true" className="mr-2 size-4" />
              Tambah Periode Baru
            </Button>
          ) : undefined
        }
      />
      <Notice text={notice} danger={error} />
      <Card className="mb-5 flex gap-3 p-4">
        <CalendarDays aria-hidden="true" className="mt-1 size-5 shrink-0 text-primary" />
        <p className="text-sm leading-6 text-body">
          Hanya satu periode dapat aktif. Mengaktifkan periode baru akan mengarsipkan periode aktif sebelumnya
          dan mempertahankan data historis.
        </p>
      </Card>
      {editing !== undefined && canManage ? (
        <PeriodEditor
          initial={editing ?? undefined}
          onCancel={() => setEditing(undefined)}
          onSaved={async () => {
            setEditing(undefined);
            setNotice("Periode berhasil disimpan.");
            await refresh();
          }}
        />
      ) : null}
      <div className="mb-4 grid gap-3 rounded-lg border border-border/70 bg-surface p-3 md:grid-cols-[minmax(16rem,1fr)_12rem]">
        <div className="relative">
          <span className="sr-only">Cari periode</span>
          <Search aria-hidden="true" className="absolute left-3 top-3 size-4 text-muted" />
          <Input
            aria-label="Cari periode"
            className="pl-9"
            placeholder="Cari nama periode atau kabinet…"
            value={search}
            onChange={(e) => {
              setPage(1);
              setSearch(e.target.value);
            }}
          />
        </div>
        <Select
          aria-label="Filter status periode"
          value={status}
          onChange={(e) => {
            setPage(1);
            setStatus(e.target.value);
          }}
        >
          <option value="">Semua Status</option>
          <option value="ACTIVE">Aktif</option>
          <option value="ARCHIVED">Diarsipkan</option>
        </Select>
      </div>
      {error && !items.length ? (
        <Alert variant="danger">Data periode tidak dapat dimuat. Periksa koneksi API lalu coba lagi.</Alert>
      ) : null}
      <div className="hidden md:block">
        <TableFrame>
          <table className="w-full text-left text-sm">
            <thead className="bg-surface-muted text-xs uppercase tracking-wide text-muted">
              <tr>
                {[
                  "Nama Periode",
                  "Tanggal Mulai",
                  "Tanggal Selesai",
                  "Jumlah Divisi",
                  "Jumlah Pengurus",
                  "Status",
                  "Aksi",
                ].map((x) => (
                  <th key={x} scope="col" className="px-4 py-3">
                    {x}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-border/70">
              {items.map((p) => (
                <tr key={p.id}>
                  <th scope="row" className="px-4 py-4 font-semibold text-primary">
                    {p.name}
                  </th>
                  <td className="px-4 py-4">{p.start_date}</td>
                  <td className="px-4 py-4">{p.end_date}</td>
                  <td className="px-4 py-4">{p.division_count}</td>
                  <td className="px-4 py-4">{p.member_count}</td>
                  <td className="px-4 py-4">
                    <Status value={p.status} />
                  </td>
                  <td className="px-4 py-4">
                    <div className="flex flex-wrap gap-2">
                      {canManage ? (
                        <>
                          <Button size="sm" variant="outline" onClick={() => setEditing(p)}>
                            Edit
                          </Button>
                          {p.status !== "ACTIVE" ? (
                            <Button size="sm" onClick={() => setPending({ period: p, action: "activate" })}>
                              Aktifkan
                            </Button>
                          ) : null}
                          {p.status !== "ARCHIVED" ? (
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => setPending({ period: p, action: "archive" })}
                            >
                              Arsipkan
                            </Button>
                          ) : null}
                        </>
                      ) : null}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </TableFrame>
      </div>
      <ul className="grid gap-3 md:hidden">
        {items.map((p) => (
          <li key={p.id}>
            <Card className="space-y-3 p-4">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <h2 className="font-heading font-bold text-primary">{p.name}</h2>
                  <p className="mt-1 text-xs text-muted">
                    {p.start_date} – {p.end_date}
                  </p>
                </div>
                <Status value={p.status} />
              </div>
              <dl className="grid grid-cols-2 gap-2 text-sm">
                <div>
                  <dt className="text-xs text-muted">Divisi</dt>
                  <dd>{p.division_count}</dd>
                </div>
                <div>
                  <dt className="text-xs text-muted">Pengurus</dt>
                  <dd>{p.member_count}</dd>
                </div>
              </dl>
              {canManage ? (
                <div className="flex flex-wrap gap-2">
                  <Button size="sm" variant="outline" onClick={() => setEditing(p)}>
                    Edit
                  </Button>
                  {p.status !== "ACTIVE" ? (
                    <Button size="sm" onClick={() => setPending({ period: p, action: "activate" })}>
                      Aktifkan
                    </Button>
                  ) : null}
                  {p.status !== "ARCHIVED" ? (
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => setPending({ period: p, action: "archive" })}
                    >
                      Arsipkan
                    </Button>
                  ) : null}
                </div>
              ) : null}
            </Card>
          </li>
        ))}
      </ul>
      {items.length === 0 && !error ? (
        <Card className="p-8 text-center text-sm text-muted">
          Belum ada periode yang cocok dengan filter.
        </Card>
      ) : null}
      {totalPages > 1 ? (
        <div className="mt-4 flex items-center justify-between text-sm">
          <span>
            Halaman {page} dari {totalPages}
          </span>
          <div className="grid w-full gap-3 md:w-auto md:grid-cols-[minmax(10rem,14rem)_auto] md:items-center">
            <Button variant="outline" disabled={page <= 1} onClick={() => setPage(page - 1)}>
              Sebelumnya
            </Button>
            <Button variant="outline" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>
              Selanjutnya
            </Button>
          </div>
        </div>
      ) : null}
      {pending ? (
        <Confirm
          title={pending.action === "activate" ? "Aktifkan periode?" : "Arsipkan periode?"}
          description={
            pending.action === "activate"
              ? "Mengaktifkan periode ini akan mengarsipkan periode aktif sebelumnya."
              : `Periode ${pending.period.name} akan diarsipkan. Data historis tetap tersimpan.`
          }
          confirmLabel={pending.action === "activate" ? "Aktifkan Periode" : "Arsipkan Periode"}
          onCancel={() => setPending(null)}
          onConfirm={() => void perform()}
        />
      ) : null}
    </div>
  );
}

type DivisionDraft = {
  period_id: number;
  name: string;
  description: string;
  status: "ACTIVE" | "INACTIVE";
  display_order: number;
  show_in_public_structure: boolean;
};
const emptyDivision: DivisionDraft = {
  period_id: 0,
  name: "",
  description: "",
  status: "ACTIVE",
  display_order: 0,
  show_in_public_structure: true,
};

export function DivisionManager() {
  const { user } = useAuth();
  const editable = user?.role !== "KEPALA_BIDANG";
  const [periods, setPeriods] = React.useState<Period[]>([]);
  const [items, setItems] = React.useState<Division[]>([]);
  const [search, setSearch] = React.useState("");
  const [periodId, setPeriodId] = React.useState("");
  const [status, setStatus] = React.useState("");
  const [draft, setDraft] = React.useState<DivisionDraft | null>(null);
  const [editing, setEditing] = React.useState<number>();
  const [notice, setNotice] = React.useState("");
  const [error, setError] = React.useState(false);
  const [confirm, setConfirm] = React.useState<Division | null>(null);
  const refresh = React.useCallback(async () => {
    try {
      const [p, d] = await Promise.all([
        getPeriods({ page: 1, limit: 100 }),
        getDivisions({ page: 1, limit: 100, search, period_id: periodId || undefined, status }),
      ]);
      setPeriods(p.items);
      setItems(d.items);
      setError(false);
    } catch {
      setError(true);
    }
  }, [search, periodId, status]);
  React.useEffect(() => {
    void refresh();
  }, [refresh]);
  async function submit(event: React.FormEvent) {
    event.preventDefault();
    if (!draft) return;
    setNotice("");
    try {
      await saveDivision(
        { ...draft, period_id: Number(draft.period_id), description: draft.description || null },
        editing,
      );
      setNotice("Data divisi berhasil disimpan.");
      setDraft(null);
      setEditing(undefined);
      await refresh();
    } catch (e) {
      setNotice(e instanceof Error ? e.message : "Data divisi gagal disimpan.");
      setError(true);
    }
  }
  async function remove() {
    if (!confirm) return;
    try {
      await deleteDivision(confirm.id);
      setNotice("Divisi berhasil dihapus.");
      setConfirm(null);
      await refresh();
    } catch (e) {
      setNotice(e instanceof Error ? e.message : "Divisi tidak dapat dihapus karena masih digunakan.");
      setError(true);
      setConfirm(null);
    }
  }
  return (
    <div className="mx-auto max-w-[1440px] p-4 md:p-6 lg:p-8">
      <AdminBreadcrumb items={[{ label: "Organisasi" }, { label: "Manajemen Divisi" }]} />
      <AdminPageHeader
        title="Manajemen Divisi"
        description="Kelola divisi berdasarkan periode kepengurusan."
        actions={
          editable ? (
            <Button
              onClick={() => {
                setDraft({
                  ...emptyDivision,
                  period_id: periods.find((p) => p.status === "ACTIVE")?.id ?? periods[0]?.id ?? 0,
                });
                setEditing(undefined);
              }}
            >
              <Plus className="mr-2 size-4" />
              Tambah Divisi
            </Button>
          ) : undefined
        }
      />
      <Notice text={notice} danger={error} />
      {draft && editable ? (
        <form
          onSubmit={(e) => void submit(e)}
          className="mb-5 grid gap-4 rounded-lg border border-border bg-surface p-4 md:grid-cols-2 md:p-5"
        >
          <div className={fieldClass}>
            <Label htmlFor="division-name">Nama Divisi</Label>
            <Input
              id="division-name"
              required
              maxLength={120}
              value={draft.name}
              onChange={(e) => setDraft({ ...draft, name: e.target.value })}
            />
          </div>
          <div className={fieldClass}>
            <Label htmlFor="division-period">Periode</Label>
            <Select
              id="division-period"
              required
              value={draft.period_id || ""}
              onChange={(e) => setDraft({ ...draft, period_id: Number(e.target.value) })}
            >
              <option value="" disabled>
                Pilih periode
              </option>
              {periods.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name} · {p.status === "ACTIVE" ? "Aktif" : "Diarsipkan"}
                </option>
              ))}
            </Select>
          </div>
          <div className={fieldClass}>
            <Label htmlFor="division-description">Deskripsi</Label>
            <Textarea
              id="division-description"
              value={draft.description}
              onChange={(e) => setDraft({ ...draft, description: e.target.value })}
            />
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className={fieldClass}>
              <Label htmlFor="division-status">Status</Label>
              <Select
                id="division-status"
                value={draft.status}
                onChange={(e) => setDraft({ ...draft, status: e.target.value as DivisionDraft["status"] })}
              >
                <option value="ACTIVE">Aktif</option>
                <option value="INACTIVE">Nonaktif</option>
              </Select>
            </div>
            <div className={fieldClass}>
              <Label htmlFor="division-order">Urutan</Label>
              <Input
                id="division-order"
                type="number"
                min={0}
                value={draft.display_order}
                onChange={(e) => setDraft({ ...draft, display_order: Number(e.target.value) })}
              />
            </div>
          </div>
          <label
            htmlFor="division-public-structure"
            className="flex min-h-11 items-start gap-3 rounded-lg border border-border/70 bg-surface-soft p-3 text-sm text-body md:col-span-2"
          >
            <Checkbox
              id="division-public-structure"
              checked={draft.show_in_public_structure}
              onChange={(e) => setDraft({ ...draft, show_in_public_structure: e.target.checked })}
              aria-describedby="division-public-structure-help"
            />
            <span>
              <span className="block font-semibold text-primary">
                Tampilkan sebagai bidang pada struktur publik
              </span>
              <span id="division-public-structure-help" className="mt-1 block text-xs leading-5 text-muted">
                Pengaturan ini hanya mengatur bagan organisasi. Divisi tetap dapat menjadi tujuan Aspirasi
                selama aktif.
              </span>
            </span>
          </label>
          <div className="flex gap-2 md:col-span-2">
            <Button type="submit">{editing ? "Simpan Perubahan" : "Simpan Divisi"}</Button>
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setDraft(null);
                setEditing(undefined);
              }}
            >
              Batal
            </Button>
          </div>
        </form>
      ) : null}
      <div className="mb-4 grid gap-3 rounded-lg border border-border/70 bg-surface p-3 md:grid-cols-3">
        <div className="relative">
          <span className="sr-only">Cari divisi</span>
          <Search aria-hidden="true" className="absolute left-3 top-3 size-4 text-muted" />
          <Input
            aria-label="Cari divisi"
            className="pl-9"
            placeholder="Cari divisi…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <Select aria-label="Filter periode" value={periodId} onChange={(e) => setPeriodId(e.target.value)}>
          <option value="">Semua Periode</option>
          {periods.map((p) => (
            <option key={p.id} value={p.id}>
              {p.name}
            </option>
          ))}
        </Select>
        <Select aria-label="Filter status" value={status} onChange={(e) => setStatus(e.target.value)}>
          <option value="">Semua Status</option>
          <option value="ACTIVE">Aktif</option>
          <option value="INACTIVE">Nonaktif</option>
        </Select>
      </div>
      {error && !items.length ? <Alert variant="danger">Data divisi tidak dapat dimuat.</Alert> : null}
      <div className="hidden md:block">
        <TableFrame>
          <table className="w-full text-left text-sm">
            <thead className="bg-surface-muted text-xs uppercase tracking-wide text-muted">
              <tr>
                {["Nama", "Kepala Bidang", "Jumlah Anggota", "Periode", "Status", "Aksi"].map((x) => (
                  <th scope="col" className="px-4 py-3" key={x}>
                    {x}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-border/70">
              {items.map((d) => (
                <tr key={d.id}>
                  <th scope="row" className="px-4 py-4 font-semibold text-primary">
                    {d.name}
                  </th>
                  <td className="px-4 py-4">{d.head?.name ?? "Belum ditetapkan"}</td>
                  <td className="px-4 py-4">{d.member_count}</td>
                  <td className="px-4 py-4">{d.period_name}</td>
                  <td className="px-4 py-4">
                    <Status value={d.status} />
                  </td>
                  <td className="px-4 py-4">
                    <div className="flex gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => {
                          setDraft({
                            period_id: d.period_id,
                            name: d.name,
                            description: d.description ?? "",
                            status: d.status,
                            display_order: d.display_order,
                            show_in_public_structure: d.show_in_public_structure,
                          });
                          setEditing(d.id);
                        }}
                        disabled={!editable}
                      >
                        Edit
                      </Button>
                      <Button size="sm" variant="outline" onClick={() => setConfirm(d)} disabled={!editable}>
                        Hapus
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </TableFrame>
      </div>
      <ul className="grid gap-3 md:hidden">
        {items.map((d) => (
          <li key={d.id}>
            <Card className="space-y-3 p-4">
              <div className="flex justify-between gap-3">
                <h2 className="font-heading font-bold text-primary">{d.name}</h2>
                <Status value={d.status} />
              </div>
              <p className="text-sm text-body">Kepala Bidang: {d.head?.name ?? "Belum ditetapkan"}</p>
              <p className="text-xs text-muted">
                {d.member_count} anggota · {d.period_name}
              </p>
              {editable ? (
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => {
                      setDraft({
                        period_id: d.period_id,
                        name: d.name,
                        description: d.description ?? "",
                        status: d.status,
                        display_order: d.display_order,
                        show_in_public_structure: d.show_in_public_structure,
                      });
                      setEditing(d.id);
                    }}
                  >
                    Edit
                  </Button>
                  <Button size="sm" variant="outline" onClick={() => setConfirm(d)}>
                    Hapus
                  </Button>
                </div>
              ) : null}
            </Card>
          </li>
        ))}
      </ul>
      {items.length === 0 && !error ? (
        <Card className="p-8 text-center text-sm text-muted">Belum ada divisi untuk filter ini.</Card>
      ) : null}
      {confirm ? (
        <Confirm
          title="Hapus divisi?"
          description={`Divisi ${confirm.name} hanya dapat dihapus jika tidak memiliki anggota atau resource terkait.`}
          confirmLabel="Hapus Divisi"
          onCancel={() => setConfirm(null)}
          onConfirm={() => void remove()}
        />
      ) : null}
    </div>
  );
}

export function MemberManager() {
  const { user } = useAuth();
  const editable = user?.role === "SUPER_ADMIN" || user?.role === "BPH";
  const [items, setItems] = React.useState<OrganizationMember[]>([]);
  const [periods, setPeriods] = React.useState<Period[]>([]);
  const [divisions, setDivisions] = React.useState<Division[]>([]);
  const [search, setSearch] = React.useState("");
  const [periodId, setPeriodId] = React.useState("");
  const [divisionId, setDivisionId] = React.useState("");
  const [position, setPosition] = React.useState("");
  const [status, setStatus] = React.useState("");
  const [notice, setNotice] = React.useState("");
  const [error, setError] = React.useState(false);
  const [confirm, setConfirm] = React.useState<OrganizationMember | null>(null);
  const refresh = React.useCallback(async () => {
    try {
      const [p, d, m] = await Promise.all([
        getPeriods({ page: 1, limit: 100 }),
        getDivisions({ page: 1, limit: 100, period_id: periodId || undefined }),
        getMembers({
          page: 1,
          limit: 100,
          search,
          period_id: periodId || undefined,
          division_id: divisionId || undefined,
          position,
          status,
        }),
      ]);
      setPeriods(p.items);
      setDivisions(d.items);
      setItems(m.items);
      setError(false);
    } catch {
      setError(true);
    }
  }, [search, periodId, divisionId, position, status]);
  React.useEffect(() => {
    void refresh();
  }, [refresh]);
  async function remove() {
    if (!confirm) return;
    try {
      await deleteMember(confirm.id);
      setNotice("Status pengurus diubah menjadi nonaktif. Riwayat tetap tersimpan.");
      setConfirm(null);
      await refresh();
    } catch (e) {
      setNotice(e instanceof Error ? e.message : "Status pengurus gagal diubah.");
      setError(true);
      setConfirm(null);
    }
  }
  const filters = (
    <div className="mb-4 grid gap-3 rounded-lg border border-border/70 bg-surface p-3 sm:grid-cols-2 xl:grid-cols-5">
      <div className="relative">
        <span className="sr-only">Cari pengurus</span>
        <Search aria-hidden="true" className="absolute left-3 top-3 size-4 text-muted" />
        <Input
          aria-label="Cari pengurus"
          className="pl-9"
          placeholder="Nama atau NIM…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>
      <Select aria-label="Filter periode" value={periodId} onChange={(e) => setPeriodId(e.target.value)}>
        <option value="">Semua Periode</option>
        {periods.map((p) => (
          <option key={p.id} value={p.id}>
            {p.name}
          </option>
        ))}
      </Select>
      <Select aria-label="Filter divisi" value={divisionId} onChange={(e) => setDivisionId(e.target.value)}>
        <option value="">Semua Divisi</option>
        {divisions.map((d) => (
          <option key={d.id} value={d.id}>
            {d.name}
          </option>
        ))}
      </Select>
      <Select aria-label="Filter jabatan" value={position} onChange={(e) => setPosition(e.target.value)}>
        <option value="">Semua Jabatan</option>
        {POSITION_OPTIONS.map((p) => (
          <option key={p.value} value={p.value}>
            {p.label}
          </option>
        ))}
      </Select>
      <Select aria-label="Filter status pengurus" value={status} onChange={(e) => setStatus(e.target.value)}>
        <option value="">Semua Status</option>
        <option value="ACTIVE">Aktif</option>
        <option value="INACTIVE">Nonaktif</option>
      </Select>
    </div>
  );
  return (
    <div className="mx-auto max-w-[1440px] p-4 md:p-6 lg:p-8">
      <AdminBreadcrumb items={[{ label: "Organisasi" }, { label: "Data Pengurus" }]} />
      <AdminPageHeader
        title="Data Pengurus"
        description="Kelola data pengurus pada setiap periode kepengurusan."
        actions={
          editable ? (
            <Button asChild>
              <Link href="/admin/pengurus/new">
                <Plus className="mr-2 size-4" />
                Tambah Pengurus
              </Link>
            </Button>
          ) : undefined
        }
      />
      <Notice text={notice} danger={error} />
      {filters}
      {error && !items.length ? <Alert variant="danger">Data pengurus tidak dapat dimuat.</Alert> : null}
      <div className="hidden md:block">
        <TableFrame>
          <table className="w-full text-left text-sm">
            <thead className="bg-surface-muted text-xs uppercase tracking-wide text-muted">
              <tr>
                {["Foto", "Nama", "NIM", "Jabatan", "Divisi", "Periode", "Status", "Aksi"].map((x) => (
                  <th scope="col" className="px-4 py-3" key={x}>
                    {x}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-border/70">
              {items.map((m) => (
                <tr key={m.id}>
                  <td className="px-4 py-3">
                    <MediaImage
                      value={m.photo_url}
                      alt={`Foto ${m.name}`}
                      width={40}
                      height={40}
                      className="size-10 rounded-full object-cover"
                      fallbackClassName="bg-primary-fixed text-primary"
                      fallbackText={memberInitials(m.name)}
                    />
                  </td>
                  <th scope="row" className="px-4 py-3 font-semibold text-primary">
                    {m.name}
                  </th>
                  <td className="px-4 py-3">{m.nim || "—"}</td>
                  <td className="px-4 py-3">{positionLabel(m.position)}</td>
                  <td className="px-4 py-3">{m.division_name || "BPH / Umum"}</td>
                  <td className="px-4 py-3">{m.period_name}</td>
                  <td className="px-4 py-3">
                    <Status value={m.status} />
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      {editable ? (
                        <>
                          <Button size="sm" variant="outline" asChild>
                            <Link href={`/admin/pengurus/${m.id}/edit`}>Edit</Link>
                          </Button>
                          <Button size="sm" variant="outline" onClick={() => setConfirm(m)}>
                            Nonaktifkan
                          </Button>
                        </>
                      ) : null}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </TableFrame>
      </div>
      <ul className="grid gap-3 md:hidden">
        {items.map((m) => (
          <li key={m.id}>
            <Card className="flex gap-3 p-4">
              <MediaImage
                value={m.photo_url}
                alt={`Foto ${m.name}`}
                width={48}
                height={48}
                className="size-12 rounded-full object-cover"
                fallbackClassName="bg-primary-fixed text-primary"
                fallbackText={memberInitials(m.name)}
              />
              <div className="min-w-0 flex-1">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <h2 className="break-words font-heading font-bold text-primary">{m.name}</h2>
                  <Status value={m.status} />
                </div>
                <p className="mt-1 text-sm">
                  {positionLabel(m.position)} · {m.division_name || "BPH / Umum"}
                </p>
                <p className="text-xs text-muted">
                  NIM {m.nim || "—"} · {m.period_name}
                </p>
                {editable ? (
                  <div className="mt-3 flex gap-2">
                    <Button size="sm" variant="outline" asChild>
                      <Link href={`/admin/pengurus/${m.id}/edit`}>Edit</Link>
                    </Button>
                    <Button size="sm" variant="outline" onClick={() => setConfirm(m)}>
                      Nonaktifkan
                    </Button>
                  </div>
                ) : null}
              </div>
            </Card>
          </li>
        ))}
      </ul>
      {items.length === 0 && !error ? (
        <Card className="p-8 text-center text-sm text-muted">Belum ada pengurus untuk filter ini.</Card>
      ) : null}
      {confirm ? (
        <Confirm
          title="Nonaktifkan pengurus?"
          description={`Data ${confirm.name} akan diubah menjadi INACTIVE dan tidak tampil di struktur publik. Riwayat tetap disimpan.`}
          confirmLabel="Nonaktifkan"
          onCancel={() => setConfirm(null)}
          onConfirm={() => void remove()}
        />
      ) : null}
    </div>
  );
}

export function MemberEditor({ memberId }: { memberId?: number }) {
  const router = useRouter();
  const { user } = useAuth();
  const canEdit = user?.role === "SUPER_ADMIN" || user?.role === "BPH";
  const [periods, setPeriods] = React.useState<Period[]>([]);
  const [divisions, setDivisions] = React.useState<Division[]>([]);
  const [draft, setDraft] = React.useState({
    period_id: 0,
    division_id: "",
    name: "",
    nim: "",
    position: "KETUA" as MemberPosition,
    photo_url: "",
    display_order: 0,
    status: "ACTIVE" as "ACTIVE" | "INACTIVE",
  });
  const [error, setError] = React.useState("");
  const [saving, setSaving] = React.useState(false);
  React.useEffect(() => {
    void Promise.all([getPeriods({ page: 1, limit: 100 }), getDivisions({ page: 1, limit: 100 })])
      .then(([p, d]) => {
        setPeriods(p.items);
        setDivisions(d.items);
        if (!memberId) {
          const active = p.items.find((x) => x.status === "ACTIVE");
          if (active) setDraft((old) => ({ ...old, period_id: active.id }));
        }
      })
      .catch(() => setError("Periode dan divisi tidak dapat dimuat."));
  }, [memberId]);
  React.useEffect(() => {
    if (memberId)
      void getMember(memberId)
        .then((m) =>
          setDraft({
            period_id: m.period_id,
            division_id: m.division_id ? String(m.division_id) : "",
            name: m.name,
            nim: m.nim ?? "",
            position: m.position,
            photo_url: m.photo_url ?? "",
            display_order: m.display_order,
            status: m.status,
          }),
        )
        .catch(() => setError("Data pengurus tidak ditemukan."));
  }, [memberId]);
  const availableDivisions = divisions.filter((d) => d.period_id === draft.period_id);
  async function submit(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    setSaving(true);
    try {
      await saveMember(
        {
          period_id: draft.period_id,
          division_id: draft.division_id ? Number(draft.division_id) : null,
          name: draft.name,
          nim: draft.nim || null,
          position: draft.position,
          photo_url: draft.photo_url || null,
          display_order: draft.display_order,
          status: draft.status,
        },
        memberId,
      );
      router.push("/admin/pengurus");
      router.refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Data pengurus gagal disimpan.");
    } finally {
      setSaving(false);
    }
  }
  const divisionRequired = draft.position === "KEPALA_BIDANG" || draft.position === "ANGGOTA";
  return (
    <div className="mx-auto max-w-5xl p-4 md:p-6 lg:p-8">
      <AdminBreadcrumb
        items={[
          { label: "Organisasi" },
          { label: "Data Pengurus", href: "/admin/pengurus" },
          { label: memberId ? "Edit Pengurus" : "Tambah Pengurus" },
        ]}
      />
      <AdminPageHeader
        title={memberId ? "Edit Data Pengurus" : "Tambah Pengurus Baru"}
        description="Lengkapi profil dan penugasan pengurus pada periode organisasi."
      />
      <Notice text={error} danger />
      {!canEdit ? (
        <Alert className="mb-4" variant="info" title="Mode baca saja">
          Anda dapat melihat detail pengurus dalam lingkup divisi, tetapi tidak dapat mengubahnya.
        </Alert>
      ) : null}
      <form
        onSubmit={(e) => void submit(e)}
        data-slot="member-form-grid"
        className="grid gap-5 lg:grid-cols-[minmax(17rem,0.85fr)_minmax(0,1.5fr)]"
      >
        <Card className="space-y-4 p-5">
          <h2 className="font-heading text-lg font-bold text-primary">Foto Profil</h2>
          <MediaUpload
            id="member-photo"
            label="Foto profil"
            purpose="MEMBER_PHOTO"
            value={draft.photo_url}
            onChange={(photo_url) => setDraft((current) => ({ ...current, photo_url }))}
            disabled={!canEdit}
            optional
          />
        </Card>
        <Card className="grid content-start gap-4 p-5 sm:grid-cols-2">
          <div className={`${fieldClass} sm:col-span-2`}>
            <Label htmlFor="member-name">Nama Lengkap</Label>
            <Input
              id="member-name"
              required
              maxLength={160}
              disabled={!canEdit}
              value={draft.name}
              onChange={(e) => setDraft({ ...draft, name: e.target.value })}
            />
          </div>
          <div className={fieldClass}>
            <Label htmlFor="member-nim">NIM (opsional)</Label>
            <Input
              id="member-nim"
              maxLength={32}
              disabled={!canEdit}
              value={draft.nim}
              onChange={(e) => setDraft({ ...draft, nim: e.target.value })}
            />
          </div>
          <div className={fieldClass}>
            <Label htmlFor="member-position">Jabatan</Label>
            <Select
              id="member-position"
              disabled={!canEdit}
              value={draft.position}
              onChange={(e) =>
                setDraft({
                  ...draft,
                  position: e.target.value as MemberPosition,
                  division_id:
                    e.target.value === "KETUA" ||
                    e.target.value === "WAKIL_KETUA" ||
                    e.target.value === "SEKRETARIS" ||
                    e.target.value === "BENDAHARA"
                      ? ""
                      : draft.division_id,
                })
              }
            >
              {POSITION_OPTIONS.map((x) => (
                <option key={x.value} value={x.value}>
                  {x.label}
                </option>
              ))}
            </Select>
          </div>
          <div className={fieldClass}>
            <Label htmlFor="member-period">Periode</Label>
            <Select
              id="member-period"
              required
              disabled={!canEdit}
              value={draft.period_id || ""}
              onChange={(e) => setDraft({ ...draft, period_id: Number(e.target.value), division_id: "" })}
            >
              <option value="" disabled>
                Pilih periode
              </option>
              {periods.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                  {p.status === "ACTIVE" ? " (Aktif)" : ""}
                </option>
              ))}
            </Select>
          </div>
          <div className={fieldClass}>
            <Label htmlFor="member-division">Divisi{divisionRequired ? " *" : " (opsional)"}</Label>
            <Select
              id="member-division"
              required={divisionRequired}
              value={draft.division_id}
              onChange={(e) => setDraft({ ...draft, division_id: e.target.value })}
              disabled={!canEdit || availableDivisions.length === 0}
            >
              <option value="">{divisionRequired ? "Pilih divisi" : "BPH / Umum"}</option>
              {availableDivisions.map((d) => (
                <option key={d.id} value={d.id}>
                  {d.name}
                </option>
              ))}
            </Select>
            {divisionRequired ? (
              <p className="text-xs text-muted">Kepala Bidang dan Anggota ditempatkan pada satu divisi.</p>
            ) : null}
          </div>
          <div className={fieldClass}>
            <Label htmlFor="member-status">Status Keanggotaan</Label>
            <Select
              id="member-status"
              disabled={!canEdit}
              value={draft.status}
              onChange={(e) => setDraft({ ...draft, status: e.target.value as "ACTIVE" | "INACTIVE" })}
            >
              <option value="ACTIVE">Aktif</option>
              <option value="INACTIVE">Nonaktif</option>
            </Select>
          </div>
          <div className={fieldClass}>
            <Label htmlFor="member-order">Urutan Tampilan</Label>
            <Input
              id="member-order"
              min={0}
              type="number"
              disabled={!canEdit}
              value={draft.display_order}
              onChange={(e) => setDraft({ ...draft, display_order: Number(e.target.value) })}
            />
          </div>
          <ActionGroup role="group" aria-label="Aksi formulir pengurus" className="sm:col-span-2">
            <Button type="submit" disabled={!canEdit || saving || !draft.period_id}>
              {saving ? "Menyimpan…" : "Simpan Pengurus"}
            </Button>
            <Button type="button" variant="outline" asChild>
              <Link href="/admin/pengurus">Batal</Link>
            </Button>
          </ActionGroup>
        </Card>
      </form>
    </div>
  );
}

export function OrganizationStructureManager() {
  const { user } = useAuth();
  const [periods, setPeriods] = React.useState<Period[]>([]);
  const [periodId, setPeriodId] = React.useState("");
  const [members, setMembers] = React.useState<OrganizationMember[]>([]);
  const [divisions, setDivisions] = React.useState<Division[]>([]);
  const [failed, setFailed] = React.useState(false);
  React.useEffect(() => {
    void getPeriods({ page: 1, limit: 100 })
      .then((data) => {
        setPeriods(data.items);
        const active = data.items.find((p) => p.status === "ACTIVE");
        setPeriodId(active ? String(active.id) : data.items[0] ? String(data.items[0].id) : "");
      })
      .catch(() => setFailed(true));
  }, []);
  React.useEffect(() => {
    if (!periodId) return;
    void Promise.all([
      getMembers({ page: 1, limit: 100, period_id: periodId, status: "ACTIVE" }),
      getDivisions({ page: 1, limit: 100, period_id: periodId, status: "ACTIVE" }),
    ])
      .then(([m, d]) => {
        setMembers(m.items);
        setDivisions(d.items);
        setFailed(false);
      })
      .catch(() => setFailed(true));
  }, [periodId]);
  const global = members.filter((m) => !m.division_id);
  const byDivision = divisions.map((d) => ({
    division: d,
    items: members.filter((m) => m.division_id === d.id),
  }));
  return (
    <div className="mx-auto max-w-[1440px] p-4 md:p-6 lg:p-8">
      <AdminBreadcrumb items={[{ label: "Organisasi" }, { label: "Struktur Organisasi" }]} />
      <AdminPageHeader
        title="Struktur Organisasi"
        description="Lihat hierarki kepengurusan dan daftar anggota berdasarkan periode."
        actions={
          <div className="flex gap-2">
            <Select
              aria-label="Pilih periode struktur"
              value={periodId}
              onChange={(e) => setPeriodId(e.target.value)}
            >
              {periods.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                  {p.status === "ACTIVE" ? " · Aktif" : ""}
                </option>
              ))}
            </Select>
            {user?.role !== "KEPALA_BIDANG" ? (
              <Button asChild>
                <Link href="/admin/pengurus/new">
                  <Plus className="mr-2 size-4" />
                  Tambah Pengurus
                </Link>
              </Button>
            ) : null}
          </div>
        }
      />
      {failed ? <Alert variant="danger">Struktur organisasi gagal dimuat.</Alert> : null}
      <section
        aria-labelledby="hierarchy-title"
        className="mb-6 rounded-xl border border-border/70 bg-primary-container/5 p-4 md:p-6"
      >
        <div className="mb-4 flex items-center gap-2">
          <Network aria-hidden="true" className="size-5 text-primary" />
          <h2 id="hierarchy-title" className="font-heading text-lg font-bold text-primary">
            Bagan Struktur
          </h2>
        </div>
        {global.length ? (
          <ul className="mb-5 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {global.map((m) => (
              <MemberHierarchyCard key={m.id} member={m} />
            ))}
          </ul>
        ) : null}
        <ul className="grid gap-4 md:grid-cols-2">
          {byDivision.map(({ division, items }) => {
            const head = items.find((m) => m.position === "KEPALA_BIDANG");
            const group = items.filter((m) => m.position === "ANGGOTA");
            return (
              <li key={division.id} className="min-w-0 rounded-xl border border-border/70 bg-surface p-4">
                <div className="flex items-start justify-between gap-3">
                  <h3 className="break-words font-heading text-base font-bold text-primary">
                    {division.name}
                  </h3>
                  <Badge variant="info">{group.length + (head ? 1 : 0)} Anggota</Badge>
                </div>
                {head ? (
                  <ul className="mt-3">
                    <MemberHierarchyCard member={head} />
                  </ul>
                ) : (
                  <p className="mt-3 rounded-md bg-warning/10 p-3 text-sm text-warning-foreground">
                    Kepala Bidang belum ditetapkan.
                  </p>
                )}
                {group.length ? (
                  <ul className="mt-2 grid gap-2 sm:grid-cols-2">
                    {group.map((m) => (
                      <MemberHierarchyCard key={m.id} member={m} />
                    ))}
                  </ul>
                ) : null}
              </li>
            );
          })}
        </ul>
        {!members.length ? (
          <p className="rounded-lg bg-surface p-6 text-center text-sm text-muted">
            Belum ada pengurus aktif pada periode ini.
          </p>
        ) : null}
      </section>
      <Card className="p-4 md:p-5">
        <div className="mb-4 flex items-center gap-2">
          <UsersRound aria-hidden="true" className="size-5 text-primary" />
          <h2 className="font-heading text-lg font-bold text-primary">Daftar Pengurus Periode</h2>
        </div>
        <ul className="divide-y divide-border/70">
          {members.map((m) => (
            <li key={m.id} className="flex flex-wrap items-center justify-between gap-3 py-3">
              <div className="min-w-0">
                <p className="break-words font-semibold text-primary">{m.name}</p>
                <p className="text-sm text-muted">
                  {positionLabel(m.position)}
                  {m.division_name ? ` · ${m.division_name}` : ""}
                </p>
              </div>
              {user?.role !== "KEPALA_BIDANG" ? (
                <Button size="sm" variant="outline" asChild>
                  <Link href={`/admin/pengurus/${m.id}/edit`}>Edit</Link>
                </Button>
              ) : null}
            </li>
          ))}
        </ul>
      </Card>
    </div>
  );
}

function MemberHierarchyCard({ member }: { member: OrganizationMember }) {
  return (
    <li className="flex min-w-0 items-center gap-3 rounded-lg border border-border/70 bg-surface p-3">
      <MediaImage
        value={member.photo_url}
        alt={`Foto ${member.name}`}
        width={40}
        height={40}
        className="size-10 rounded-full object-cover"
        fallbackClassName="bg-primary-fixed text-primary"
        fallbackText={memberInitials(member.name)}
      />
      <span className="min-w-0">
        <span className="block text-xs font-semibold text-secondary">{positionLabel(member.position)}</span>
        <span className="block break-words text-sm font-bold text-primary">{member.name}</span>
      </span>
    </li>
  );
}

function memberInitials(name: string) {
  return name
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? "")
    .join("");
}

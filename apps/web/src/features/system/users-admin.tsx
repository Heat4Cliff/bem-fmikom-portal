"use client";

import { MoreHorizontal, Plus, Search, ShieldCheck, UserRoundCog } from "lucide-react";
import * as React from "react";
import { AdminPageHeader } from "@/components/admin/admin-page-header";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogTitle } from "@/components/ui/overlays";
import { Badge, Input, Label, Select } from "@/components/ui/primitives";
import { Alert } from "@/components/ui/primitives";
import { getDivisions } from "@/features/organization/api";
import type { Division } from "@/features/organization/types";
import { useAuth } from "@/lib/auth/auth-provider";
import { ApiError } from "@/lib/api/client";
import {
  createUser,
  deactivateUser,
  getUsers,
  reactivateUser,
  resetUserPassword,
  updateUser,
  type ManagedUser,
} from "./api";

type Editor = {
  id?: number;
  name: string;
  email: string;
  role: ManagedUser["role"];
  division_id: string;
  status: ManagedUser["status"];
  password: string;
};
const blank: Editor = { name: "", email: "", role: "BPH", division_id: "", status: "ACTIVE", password: "" };
const roleName = (role: string) =>
  role === "SUPER_ADMIN" ? "Super Admin" : role === "BPH" ? "BPH" : "Kepala Bidang";
const date = (value: string | null) =>
  value
    ? new Intl.DateTimeFormat("id-ID", {
        dateStyle: "medium",
        timeStyle: "short",
        timeZone: "Asia/Jakarta",
      }).format(new Date(value))
    : "Belum pernah";

export function UsersAdmin() {
  const { user } = useAuth();
  const [items, setItems] = React.useState<ManagedUser[]>([]);
  const [divisions, setDivisions] = React.useState<Division[]>([]);
  const [editor, setEditor] = React.useState<Editor | null>(null);
  const [resetTarget, setResetTarget] = React.useState<ManagedUser | null>(null);
  const [reset, setReset] = React.useState({ password: "", confirm: "" });
  const [page, setPage] = React.useState(1);
  const [pages, setPages] = React.useState(1);
  const [search, setSearch] = React.useState("");
  const [role, setRole] = React.useState("");
  const [division, setDivision] = React.useState("");
  const [status, setStatus] = React.useState("");
  const [loading, setLoading] = React.useState(true);
  const [busy, setBusy] = React.useState(false);
  const [error, setError] = React.useState("");
  const [message, setMessage] = React.useState("");
  const allowed = user?.role === "SUPER_ADMIN";
  const load = React.useCallback(async () => {
    if (!allowed) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError("");
    try {
      const r = await getUsers({ page, limit: 20, search, role, division_id: division, status });
      setItems(r.items);
      setPages(Math.max(1, r.meta.total_pages));
    } catch {
      setError("Daftar pengguna tidak dapat dimuat.");
    } finally {
      setLoading(false);
    }
  }, [allowed, page, search, role, division, status]);
  React.useEffect(() => {
    void load();
  }, [load]);
  React.useEffect(() => {
    if (allowed)
      void getDivisions({ limit: 100 })
        .then((r) => setDivisions(r.items))
        .catch(() => setDivisions([]));
  }, [allowed]);
  const act = async (fn: () => Promise<unknown>, success: string) => {
    setBusy(true);
    setError("");
    setMessage("");
    try {
      await fn();
      setMessage(success);
      await load();
    } catch (e) {
      setError(
        e instanceof ApiError && e.status === 409
          ? "Operasi ditolak untuk melindungi akses administrator atau akun Anda."
          : "Permintaan gagal. Periksa data dan hak akses.",
      );
    } finally {
      setBusy(false);
    }
  };
  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editor) return;
    const payload = {
      name: editor.name,
      email: editor.email,
      role: editor.role,
      division_id: editor.role === "KEPALA_BIDANG" ? Number(editor.division_id) : null,
      status: editor.status,
      ...(!editor.id ? { password: editor.password } : {}),
    };
    await act(
      () => (editor.id ? updateUser(editor.id, payload) : createUser(payload)),
      editor.id ? "Data pengguna diperbarui." : "Pengguna berhasil dibuat.",
    );
    setEditor(null);
  };
  const roleChange = (next: ManagedUser["role"]) =>
    setEditor((current) =>
      current
        ? { ...current, role: next, division_id: next === "KEPALA_BIDANG" ? current.division_id : "" }
        : current,
    );
  const actions = (item: ManagedUser) => (
    <details className="relative">
      <summary
        aria-label={`Aksi untuk ${item.name}`}
        className="grid size-10 cursor-pointer list-none place-items-center rounded-md hover:bg-surface-muted focus-visible:outline-2"
      >
        <MoreHorizontal className="size-5" />
      </summary>
      <div className="absolute right-0 z-10 mt-1 grid min-w-44 rounded-lg border border-border bg-surface p-1 shadow-float">
        <button
          type="button"
          className="min-h-10 px-3 text-left text-sm hover:bg-surface-muted"
          onClick={() =>
            setEditor({
              id: item.id,
              name: item.name,
              email: item.email,
              role: item.role,
              division_id: item.division_id ? String(item.division_id) : "",
              status: item.status,
              password: "",
            })
          }
        >
          Edit pengguna
        </button>
        {item.id !== user?.id && (
          <button
            type="button"
            className="min-h-10 px-3 text-left text-sm hover:bg-surface-muted"
            onClick={() =>
              void act(
                () => (item.status === "ACTIVE" ? deactivateUser(item.id) : reactivateUser(item.id)),
                item.status === "ACTIVE" ? "Pengguna dinonaktifkan." : "Pengguna diaktifkan kembali.",
              )
            }
          >
            {item.status === "ACTIVE" ? "Nonaktifkan" : "Aktifkan kembali"}
          </button>
        )}
        <button
          type="button"
          className="min-h-10 px-3 text-left text-sm hover:bg-surface-muted"
          onClick={() => {
            setResetTarget(item);
            setReset({ password: "", confirm: "" });
          }}
        >
          Reset password
        </button>
      </div>
    </details>
  );
  if (!allowed)
    return (
      <div className="p-6">
        <Alert variant="danger" title="Akses ditolak">
          Manajemen pengguna hanya tersedia untuk SUPER_ADMIN.
        </Alert>
      </div>
    );
  return (
    <div className="space-y-5 p-4 md:p-6 lg:p-8">
      <AdminPageHeader
        title="Manajemen Pengguna & Hak Akses"
        description="Kelola akun, role, divisi, dan status login administrator."
        actions={
          <Button onClick={() => setEditor({ ...blank })}>
            <Plus className="size-4" />
            Tambah Pengguna
          </Button>
        }
      />
      <div className="flex items-center gap-2 rounded-lg border border-primary/15 bg-primary/5 p-3 text-sm text-primary">
        <ShieldCheck className="size-5 shrink-0" />
        <span>Pengelolaan akun dibatasi untuk SUPER_ADMIN. Password tidak pernah ditampilkan kembali.</span>
      </div>
      {error && (
        <p role="alert" className="rounded-md bg-danger/10 p-3 text-sm text-danger">
          {error}
        </p>
      )}
      {message && (
        <p role="status" className="rounded-md bg-success/10 p-3 text-sm text-emerald-900">
          {message}
        </p>
      )}
      <section
        aria-label="Filter pengguna"
        className="grid gap-2 rounded-xl border border-border bg-surface p-3 sm:grid-cols-2 xl:grid-cols-[minmax(14rem,1fr)_repeat(3,minmax(10rem,.55fr))_auto]"
      >
        <div className="relative">
          <span className="sr-only">Cari nama atau email</span>
          <Search className="absolute left-3 top-3 size-4 text-muted" />
          <Input
            aria-label="Cari nama atau email"
            className="pl-9"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
            placeholder="Cari nama atau email"
          />
        </div>
        <div>
          <span className="sr-only">Filter role</span>
          <Select
            aria-label="Filter role"
            value={role}
            onChange={(e) => {
              setRole(e.target.value);
              setPage(1);
            }}
          >
            <option value="">Semua role</option>
            <option value="SUPER_ADMIN">SUPER_ADMIN</option>
            <option value="BPH">BPH</option>
            <option value="KEPALA_BIDANG">KEPALA_BIDANG</option>
          </Select>
        </div>
        <div>
          <span className="sr-only">Filter divisi</span>
          <Select
            aria-label="Filter divisi"
            value={division}
            onChange={(e) => {
              setDivision(e.target.value);
              setPage(1);
            }}
          >
            <option value="">Semua divisi</option>
            {divisions.map((d) => (
              <option key={d.id} value={d.id}>
                {d.name}
              </option>
            ))}
          </Select>
        </div>
        <div>
          <span className="sr-only">Filter status</span>
          <Select
            aria-label="Filter status"
            value={status}
            onChange={(e) => {
              setStatus(e.target.value);
              setPage(1);
            }}
          >
            <option value="">Semua status</option>
            <option value="ACTIVE">Aktif</option>
            <option value="INACTIVE">Nonaktif</option>
          </Select>
        </div>
        <Button
          variant="outline"
          onClick={() => {
            setSearch("");
            setRole("");
            setDivision("");
            setStatus("");
            setPage(1);
          }}
        >
          Reset filter
        </Button>
      </section>
      <section
        aria-label="Daftar pengguna"
        className="overflow-hidden rounded-xl border border-border bg-surface"
      >
        {loading ? (
          <p className="p-6 text-sm text-muted" role="status">
            Memuat pengguna…
          </p>
        ) : items.length === 0 ? (
          <div className="grid justify-items-center gap-2 p-10 text-center">
            <UserRoundCog className="size-8 text-muted" />
            <p className="font-semibold text-primary">Tidak ada pengguna ditemukan</p>
            <p className="text-sm text-muted">Sesuaikan filter atau tambahkan pengguna baru.</p>
          </div>
        ) : (
          <>
            <div className="hidden overflow-x-auto md:block">
              <table className="w-full min-w-[900px] text-left text-sm">
                <thead className="bg-surface-muted text-xs uppercase tracking-wide text-muted">
                  <tr>
                    {["Nama & email", "Role", "Divisi / scope", "Status", "Login terakhir", "Aksi"].map(
                      (x) => (
                        <th className="px-4 py-3" key={x}>
                          {x}
                        </th>
                      ),
                    )}
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {items.map((item) => (
                    <tr key={item.id} className="hover:bg-surface-muted/60">
                      <td className="px-4 py-3">
                        <p className="font-semibold text-primary">
                          {item.name}
                          {item.id === user?.id && <span className="ml-2 text-xs text-muted">Anda</span>}
                        </p>
                        <p className="text-xs text-muted">{item.email}</p>
                      </td>
                      <td className="px-4 py-3">
                        <Badge>{roleName(item.role)}</Badge>
                      </td>
                      <td className="px-4 py-3">{item.division?.name ?? "Semua divisi"}</td>
                      <td className="px-4 py-3">
                        <Badge variant={item.status === "ACTIVE" ? "success" : "neutral"}>
                          {item.status === "ACTIVE" ? "Aktif" : "Nonaktif"}
                        </Badge>
                      </td>
                      <td className="px-4 py-3 text-xs text-muted">{date(item.last_login_at)}</td>
                      <td className="px-4 py-3">{actions(item)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <div className="grid gap-3 p-3 md:hidden">
              {items.map((item) => (
                <article key={item.id} className="rounded-lg border border-border p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <h2 className="truncate font-semibold text-primary">
                        {item.name}
                        {item.id === user?.id && " (Anda)"}
                      </h2>
                      <p className="break-all text-sm text-muted">{item.email}</p>
                    </div>
                    {actions(item)}
                  </div>
                  <dl className="mt-3 grid grid-cols-2 gap-3 text-sm">
                    <div>
                      <dt className="text-xs text-muted">Role</dt>
                      <dd>
                        <Badge>{roleName(item.role)}</Badge>
                      </dd>
                    </div>
                    <div>
                      <dt className="text-xs text-muted">Status</dt>
                      <dd>
                        <Badge variant={item.status === "ACTIVE" ? "success" : "neutral"}>
                          {item.status}
                        </Badge>
                      </dd>
                    </div>
                    <div>
                      <dt className="text-xs text-muted">Divisi</dt>
                      <dd>{item.division?.name ?? "Semua divisi"}</dd>
                    </div>
                    <div>
                      <dt className="text-xs text-muted">Login terakhir</dt>
                      <dd>{date(item.last_login_at)}</dd>
                    </div>
                  </dl>
                </article>
              ))}
            </div>
            <div className="flex flex-wrap items-center justify-between gap-3 border-t border-border bg-surface-muted px-4 py-3 text-sm">
              <span>
                Halaman {page} dari {pages}
              </span>
              <div className="flex gap-2">
                <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>
                  Sebelumnya
                </Button>
                <Button
                  variant="outline"
                  size="sm"
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
      <Dialog open={Boolean(editor)} onOpenChange={(open) => !open && setEditor(null)}>
        <DialogContent>
          <DialogTitle>{editor?.id ? "Edit Pengguna" : "Tambah Pengguna"}</DialogTitle>
          <DialogDescription>
            Field bertanda wajib harus dilengkapi. Password hanya diminta saat akun dibuat.
          </DialogDescription>
          {editor && (
            <form className="mt-5 grid gap-4" onSubmit={submit}>
              <div className="grid gap-1.5">
                <Label htmlFor="user-name">Nama Lengkap</Label>
                <Input
                  id="user-name"
                  required
                  maxLength={160}
                  value={editor.name}
                  onChange={(e) => setEditor({ ...editor, name: e.target.value })}
                />
              </div>
              <div className="grid gap-1.5">
                <Label htmlFor="user-email">Email</Label>
                <Input
                  id="user-email"
                  required
                  type="email"
                  maxLength={254}
                  value={editor.email}
                  onChange={(e) => setEditor({ ...editor, email: e.target.value })}
                />
              </div>
              <div className="grid gap-1.5">
                <Label htmlFor="user-role">Role</Label>
                <Select
                  id="user-role"
                  value={editor.role}
                  onChange={(e) => roleChange(e.target.value as ManagedUser["role"])}
                >
                  <option value="SUPER_ADMIN">SUPER_ADMIN</option>
                  <option value="BPH">BPH</option>
                  <option value="KEPALA_BIDANG">KEPALA_BIDANG</option>
                </Select>
              </div>
              {editor.role === "KEPALA_BIDANG" && (
                <div className="grid gap-1.5">
                  <Label htmlFor="user-division">Divisi</Label>
                  <Select
                    id="user-division"
                    required
                    value={editor.division_id}
                    onChange={(e) => setEditor({ ...editor, division_id: e.target.value })}
                  >
                    <option value="">Pilih divisi aktif</option>
                    {divisions.map((d) => (
                      <option key={d.id} value={d.id}>
                        {d.name}
                      </option>
                    ))}
                  </Select>
                </div>
              )}
              <div className="grid gap-1.5">
                <Label htmlFor="user-status">Status</Label>
                <Select
                  id="user-status"
                  value={editor.status}
                  onChange={(e) => setEditor({ ...editor, status: e.target.value as ManagedUser["status"] })}
                >
                  <option value="ACTIVE">Aktif</option>
                  <option value="INACTIVE">Nonaktif</option>
                </Select>
              </div>
              {!editor.id && (
                <div className="grid gap-1.5">
                  <Label htmlFor="user-temp-password">Password sementara (min. 12 karakter)</Label>
                  <Input
                    id="user-temp-password"
                    type="password"
                    autoComplete="new-password"
                    required
                    minLength={12}
                    value={editor.password}
                    onChange={(e) => setEditor({ ...editor, password: e.target.value })}
                  />
                </div>
              )}
              <div className="flex justify-end gap-2">
                <Button type="button" variant="outline" onClick={() => setEditor(null)}>
                  Batal
                </Button>
                <Button type="submit" disabled={busy}>
                  {busy ? "Menyimpan…" : "Simpan"}
                </Button>
              </div>
            </form>
          )}
        </DialogContent>
      </Dialog>
      <Dialog open={Boolean(resetTarget)} onOpenChange={(open) => !open && setResetTarget(null)}>
        <DialogContent>
          <DialogTitle>Reset password</DialogTitle>
          <DialogDescription>
            Password pengguna akan direset. Sampaikan password baru melalui kanal yang aman. Sesi refresh
            pengguna akan dicabut.
          </DialogDescription>
          <form
            className="mt-5 grid gap-4"
            onSubmit={async (e) => {
              e.preventDefault();
              if (!resetTarget) return;
              await act(
                () => resetUserPassword(resetTarget.id, reset.password, reset.confirm),
                "Password direset dan sesi refresh lama dicabut.",
              );
              setResetTarget(null);
            }}
          >
            <div className="grid gap-1.5">
              <Label htmlFor="reset-password">Password baru sementara</Label>
              <Input
                id="reset-password"
                type="password"
                autoComplete="new-password"
                minLength={12}
                required
                value={reset.password}
                onChange={(e) => setReset({ ...reset, password: e.target.value })}
              />
            </div>
            <div className="grid gap-1.5">
              <Label htmlFor="reset-confirm">Konfirmasi password</Label>
              <Input
                id="reset-confirm"
                type="password"
                autoComplete="new-password"
                minLength={12}
                required
                value={reset.confirm}
                onChange={(e) => setReset({ ...reset, confirm: e.target.value })}
              />
            </div>
            <div className="flex justify-end gap-2">
              <Button type="button" variant="outline" onClick={() => setResetTarget(null)}>
                Batal
              </Button>
              <Button type="submit" disabled={busy || reset.password !== reset.confirm}>
                Reset Password
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}

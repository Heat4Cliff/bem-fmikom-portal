import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { AdminShell } from "@/components/admin/admin-shell";
import { apiRequest, getMemoryAccessToken, onAuthExpired, setMemoryAccessToken } from "@/lib/api/client";
import { AdminRouteGuard, AuthProvider, useAuth } from "@/lib/auth/auth-provider";
import AdminLoginPage from "@/app/(admin)/admin/login/page";
import { validateLogin } from "@/lib/auth/login-validation";
import type { AuthUser } from "@/types/auth";

const navigation = vi.hoisted(() => ({ pathname: "/admin/login", replace: vi.fn() }));
vi.mock("next/navigation", () => ({
  usePathname: () => navigation.pathname,
  useRouter: () => ({ replace: navigation.replace, push: vi.fn(), refresh: vi.fn() }),
}));

const activeUser: AuthUser = {
  id: 7,
  name: "Pengurus BEM",
  email: "pengurus@example.org",
  role: "BPH",
  division: null,
};

function jsonResponse(status: number, body: unknown) {
  return new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
}

function payload<T>(data: T) {
  return { success: true, message: "ok", data };
}

function AuthProbe() {
  const { status, user, logout } = useAuth();
  return (
    <div>
      <span data-testid="auth-status">{status}</span>
      <span data-testid="auth-user">{user?.email ?? "none"}</span>
      <button type="button" onClick={() => void logout()}>
        keluar
      </button>
    </div>
  );
}

describe("Phase 2 authentication UI and API client", () => {
  beforeEach(() => {
    navigation.pathname = "/admin/login";
    navigation.replace.mockClear();
    setMemoryAccessToken(null);
    onAuthExpired(null);
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("validates login email and required fields", () => {
    expect(validateLogin("", "x")).toEqual({ email: "Email wajib diisi" });
    expect(validateLogin("nope", "x")).toEqual({ email: "Masukkan alamat email yang valid" });
    expect(validateLogin("admin@example.org", "")).toEqual({ password: "Password wajib diisi" });
    expect(validateLogin(" Admin@Example.org ", "password")).toEqual({});
  });

  it("shows field errors and a pending/error state for login", async () => {
    const user = userEvent.setup();
    let releaseLogin: ((response: Response) => void) | undefined;
    const fetchMock = vi.fn((input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith("/auth/refresh"))
        return Promise.resolve(
          jsonResponse(401, { success: false, message: "expired", code: "UNAUTHENTICATED" }),
        );
      return new Promise<Response>((resolve) => {
        releaseLogin = resolve;
      });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(
      <AuthProvider>
        <AdminLoginPage />
      </AuthProvider>,
    );

    const submit = await screen.findByRole("button", { name: "Masuk" });
    await user.click(submit);
    expect(await screen.findByText("Email wajib diisi")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledTimes(1);

    await user.type(screen.getByLabelText("Email"), "admin@example.org");
    await user.type(screen.getByLabelText("Password"), "wrong password");
    await user.click(submit);
    expect(submit).toBeDisabled();
    releaseLogin?.(
      jsonResponse(401, {
        success: false,
        message: "Email atau password tidak valid",
        code: "UNAUTHENTICATED",
      }),
    );
    expect(await screen.findByText("Email atau password tidak valid")).toBeInTheDocument();
    expect(submit).toBeEnabled();
  });

  it("bootstraps current identity by refreshing the cookie and then calling auth/me", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith("/auth/refresh"))
        return jsonResponse(
          200,
          payload({ access_token: "memory-token", expires_in: 900, user: activeUser }),
        );
      if (url.endsWith("/auth/me")) return jsonResponse(200, payload(activeUser));
      return jsonResponse(500, { success: false, message: "unexpected" });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(
      <AuthProvider>
        <AuthProbe />
      </AuthProvider>,
    );
    await waitFor(() => expect(screen.getByTestId("auth-status")).toHaveTextContent("authenticated"));
    expect(screen.getByTestId("auth-user")).toHaveTextContent(activeUser.email);
    expect(getMemoryAccessToken()).toBe("memory-token");
    expect(fetchMock.mock.calls.map(([input]) => String(input))).toEqual([
      "http://localhost:8080/api/v1/auth/refresh",
      "http://localhost:8080/api/v1/auth/me",
    ]);
  });

  it("redirects unauthenticated admin routes to login after bootstrap", async () => {
    navigation.pathname = "/admin/users";
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => jsonResponse(401, { success: false, message: "expired" })),
    );
    render(
      <AuthProvider>
        <AdminRouteGuard>
          <div>protected</div>
        </AdminRouteGuard>
      </AuthProvider>,
    );
    await waitFor(() => expect(navigation.replace).toHaveBeenCalledWith("/admin/login"));
    expect(screen.queryByText("protected")).not.toBeInTheDocument();
  });

  it("retries an authenticated request once after refresh and keeps the access token in memory", async () => {
    setMemoryAccessToken("old-token");
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse(401, { success: false, message: "expired", code: "UNAUTHENTICATED" }),
      )
      .mockResolvedValueOnce(
        jsonResponse(200, payload({ access_token: "new-token", expires_in: 900, user: activeUser })),
      )
      .mockResolvedValueOnce(jsonResponse(200, payload({ ok: true })));
    vi.stubGlobal("fetch", fetchMock);

    await expect(apiRequest<{ ok: boolean }>("/admin/probe")).resolves.toEqual({ ok: true });
    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(new Headers(fetchMock.mock.calls[0][1]?.headers).get("Authorization")).toBe("Bearer old-token");
    expect(new Headers(fetchMock.mock.calls[2][1]?.headers).get("Authorization")).toBe("Bearer new-token");
    expect(getMemoryAccessToken()).toBe("new-token");
  });

  it("clears in-memory identity after logout", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith("/auth/refresh"))
        return jsonResponse(
          200,
          payload({ access_token: "memory-token", expires_in: 900, user: activeUser }),
        );
      if (url.endsWith("/auth/me")) return jsonResponse(200, payload(activeUser));
      return jsonResponse(200, payload(null));
    });
    vi.stubGlobal("fetch", fetchMock);
    render(
      <AuthProvider>
        <AuthProbe />
      </AuthProvider>,
    );
    await screen.findByText(activeUser.email);
    await userEvent.setup().click(screen.getByRole("button", { name: "keluar" }));
    await waitFor(() => expect(screen.getByTestId("auth-status")).toHaveTextContent("unauthenticated"));
    expect(getMemoryAccessToken()).toBeNull();
  });

  it("shows the current division scope only for KEPALA_BIDANG", () => {
    const kepala = { ...activeUser, role: "KEPALA_BIDANG" as const, division: { id: 3, name: "Akademik" } };
    const { rerender } = render(
      <AdminShell user={kepala} context={{ role: kepala.role, divisionId: 3 }} onLogout={() => {}}>
        <div>content</div>
      </AdminShell>,
    );
    expect(screen.queryByRole("link", { name: "Manajemen Pengguna" })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Activity Log" })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Pengaturan Landing Page" })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Berita" })).toBeInTheDocument();
    expect(screen.getByText("Scope: Akademik")).toBeInTheDocument();

    const kepalaKominfo = { ...kepala, division: { id: 4, name: "Kominfo" } };
    rerender(
      <AdminShell
        user={kepalaKominfo}
        context={{ role: kepalaKominfo.role, divisionId: 4 }}
        onLogout={() => {}}
      >
        <div>content</div>
      </AdminShell>,
    );
    expect(screen.getByText("Scope: Kominfo")).toBeInTheDocument();
    expect(screen.queryByText("Scope: Akademik")).not.toBeInTheDocument();

    rerender(
      <AdminShell user={activeUser} context={{ role: "BPH" }} onLogout={() => {}}>
        <div>content</div>
      </AdminShell>,
    );
    expect(screen.queryByRole("link", { name: "Manajemen Pengguna" })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Activity Log" })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Pengaturan Landing Page" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Periode Kepengurusan" })).toBeInTheDocument();
    expect(screen.queryByText(/^Scope:/)).not.toBeInTheDocument();

    const superAdmin = { ...activeUser, role: "SUPER_ADMIN" as const, division: null };
    rerender(
      <AdminShell user={superAdmin} context={{ role: superAdmin.role }} onLogout={() => {}}>
        <div>content</div>
      </AdminShell>,
    );
    expect(screen.queryByText(/^Scope:/)).not.toBeInTheDocument();
  });

  it("uses a safe scope fallback when a KEPALA_BIDANG division is unavailable", () => {
    const kepalaWithoutDivision = {
      ...activeUser,
      role: "KEPALA_BIDANG" as const,
      division: null,
    };
    render(
      <AdminShell
        user={kepalaWithoutDivision}
        context={{ role: kepalaWithoutDivision.role, divisionId: null }}
        onLogout={() => {}}
      >
        <div>content</div>
      </AdminShell>,
    );

    expect(screen.getByText("Scope divisi tidak tersedia")).toBeInTheDocument();
    expect(screen.queryByText("Scope: Semua Divisi")).not.toBeInTheDocument();
  });
});

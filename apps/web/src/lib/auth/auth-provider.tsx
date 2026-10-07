"use client";

import * as React from "react";
import { usePathname, useRouter } from "next/navigation";
import { AdminShell } from "@/components/admin/admin-shell";
import {
  getCurrentUser,
  loginRequest,
  logoutRequest,
  onAuthExpired,
  refreshAccessToken,
  setMemoryAccessToken,
} from "@/lib/api/client";
import type { AuthUser } from "@/types/auth";

type AuthStatus = "loading" | "authenticated" | "unauthenticated";
interface AuthContextValue {
  user: AuthUser | null;
  status: AuthStatus;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = React.createContext<AuthContextValue | null>(null);

export function useAuth() {
  const value = React.useContext(AuthContext);
  if (!value) throw new Error("useAuth must be used inside AuthProvider");
  return value;
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = React.useState<AuthUser | null>(null);
  const [status, setStatus] = React.useState<AuthStatus>("loading");

  React.useEffect(() => {
    onAuthExpired(() => {
      setMemoryAccessToken(null);
      setUser(null);
      setStatus("unauthenticated");
    });
    let active = true;
    void (async () => {
      try {
        await refreshAccessToken();
        const currentUser = await getCurrentUser();
        if (active) {
          setUser(currentUser);
          setStatus("authenticated");
        }
      } catch {
        setMemoryAccessToken(null);
        if (active) {
          setUser(null);
          setStatus("unauthenticated");
        }
      }
    })();
    return () => {
      active = false;
      onAuthExpired(null);
    };
  }, []);

  const login = React.useCallback(async (email: string, password: string) => {
    const result = await loginRequest(email, password);
    setMemoryAccessToken(result.access_token);
    setUser(result.user);
    setStatus("authenticated");
  }, []);

  const logout = React.useCallback(async () => {
    try {
      await logoutRequest();
    } finally {
      setMemoryAccessToken(null);
      setUser(null);
      setStatus("unauthenticated");
    }
  }, []);

  const value = React.useMemo(() => ({ user, status, login, logout }), [user, status, login, logout]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function AdminRouteGuard({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const { user, status, logout } = useAuth();
  const isLoginRoute = pathname === "/admin/login";

  React.useEffect(() => {
    if (status === "authenticated" && isLoginRoute) router.replace("/admin");
    if (status === "unauthenticated" && !isLoginRoute) router.replace("/admin/login");
  }, [isLoginRoute, router, status]);

  if (status === "loading") return <AuthLoading />;
  if (isLoginRoute) return status === "unauthenticated" ? children : <AuthLoading />;
  if (status !== "authenticated" || !user) return <AuthLoading />;
  return (
    <AdminShell
      context={{ role: user.role, divisionId: user.division?.id ?? null }}
      user={user}
      onLogout={() => void logout()}
    >
      {children}
    </AdminShell>
  );
}

export function AuthLoading() {
  return (
    <main
      className="grid min-h-dvh place-items-center bg-background px-4"
      aria-live="polite"
      aria-busy="true"
    >
      <div
        role="status"
        className="w-full max-w-sm rounded-lg border border-border bg-surface p-6 shadow-card"
      >
        <span className="sr-only">Memeriksa sesi administrator</span>
        <div className="mx-auto mb-4 size-10 animate-pulse rounded-lg bg-primary/15" />
        <div className="mx-auto h-4 w-40 animate-pulse rounded bg-surface-muted" />
        <div className="mx-auto mt-3 h-3 w-56 animate-pulse rounded bg-surface-muted" />
      </div>
    </main>
  );
}

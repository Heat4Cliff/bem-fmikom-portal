"use client";

import * as React from "react";
import { useRouter } from "next/navigation";
import { Eye, EyeOff, ShieldCheck } from "lucide-react";
import { Alert, Card, Input, Label } from "@/components/ui/primitives";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/api/client";
import { useAuth } from "@/lib/auth/auth-provider";
import { validateLogin } from "@/lib/auth/login-validation";
import { ThemeSwitcher } from "@/components/ui/theme-switcher";
import { BrandLogo } from "@/components/ui/brand-logo";
import { useBranding } from "@/features/content/branding-context";

export default function AdminLoginPage() {
  const { login } = useAuth();
  const branding = useBranding();
  const router = useRouter();
  const [email, setEmail] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [showPassword, setShowPassword] = React.useState(false);
  const [errors, setErrors] = React.useState<{ email?: string; password?: string }>({});
  const [serverError, setServerError] = React.useState("");
  const [loading, setLoading] = React.useState(false);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setServerError("");
    const validation = validateLogin(email, password);
    setErrors(validation);
    if (Object.keys(validation).length > 0) return;

    setLoading(true);
    try {
      await login(email.trim().toLowerCase(), password);
      router.replace("/admin");
    } catch (error) {
      setServerError(
        error instanceof ApiError ? error.message : "Tidak dapat menghubungi server. Coba lagi.",
      );
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="grid min-h-dvh place-items-center bg-background px-4 py-10">
      <div className="w-full max-w-[27rem]">
        <div className="mb-4 flex justify-end">
          <ThemeSwitcher />
        </div>
        <div className="mb-6 text-center">
          <BrandLogo
            src={branding.logoUrl}
            shape={branding.logoShape}
            fit={branding.logoFit}
            alt="BEM FMIKOM UNUGHA"
            className="mx-auto mb-4 size-14 shadow-card"
          />
          <p className="font-heading text-xl font-bold text-primary">BEM FMIKOM UNUGHA</p>
          <p className="mt-1 text-sm text-muted">Portal pengelolaan organisasi</p>
        </div>

        <Card className="p-6 shadow-card sm:p-8">
          <div className="mb-6 flex items-center gap-3">
            <span className="grid size-10 place-items-center rounded-lg bg-primary-container/10 text-primary">
              <ShieldCheck aria-hidden="true" className="size-5" />
            </span>
            <div>
              <h1 className="font-heading text-lg font-bold text-primary">Masuk ke portal</h1>
              <p className="text-sm text-muted">Gunakan akun administrator BEM.</p>
            </div>
          </div>

          {serverError ? (
            <Alert variant="danger" className="mb-5">
              {serverError}
            </Alert>
          ) : null}

          <form onSubmit={(event) => void handleSubmit(event)} noValidate className="space-y-5">
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                autoComplete="username"
                inputMode="email"
                maxLength={254}
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                placeholder="nama@bem.unugha.ac.id"
                error={errors.email}
                disabled={loading}
                required
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <div className="relative">
                <Input
                  id="password"
                  type={showPassword ? "text" : "password"}
                  autoComplete="current-password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  placeholder="Masukkan password"
                  error={errors.password}
                  disabled={loading}
                  required
                  className="pr-12"
                />
                <button
                  type="button"
                  aria-label={showPassword ? "Sembunyikan password" : "Tampilkan password"}
                  aria-pressed={showPassword}
                  onClick={() => setShowPassword((visible) => !visible)}
                  className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-2 text-muted hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-secondary"
                >
                  {showPassword ? (
                    <EyeOff aria-hidden="true" className="size-4" />
                  ) : (
                    <Eye aria-hidden="true" className="size-4" />
                  )}
                </button>
              </div>
            </div>

            <Button type="submit" loading={loading} className="w-full">
              Masuk
            </Button>
          </form>

          <p className="mt-6 text-center text-xs leading-5 text-muted">
            Akses terbatas untuk pengurus yang memiliki akun aktif.
          </p>
        </Card>
        <p className="mt-5 text-center text-xs text-muted">
          BEM FMIKOM · Universitas Nahdlatul Ulama Al Ghazali
        </p>
      </div>
    </main>
  );
}

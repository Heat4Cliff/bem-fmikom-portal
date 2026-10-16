"use client";

import * as React from "react";
import { Check, Palette } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/overlays";

export const BRAND_THEMES = ["nawasena", "classic", "salt", "matcha", "mint"] as const;
export type BrandTheme = (typeof BRAND_THEMES)[number];

const THEME_EVENT = "bem-theme-change";
const THEME_STORAGE_KEY = "bem-fmikom-theme";
const LEGACY_THEME_STORAGE_KEY = "bem-theme";

const themes: {
  id: BrandTheme;
  name: string;
  description: string;
  swatches: [string, string, string, string];
}[] = [
  {
    id: "nawasena",
    name: "Nawasena",
    description: "Gelap & elegan",
    swatches: ["#090806", "#1c1912", "#fac90a", "#ffffff"],
  },
  {
    id: "classic",
    name: "Classic",
    description: "Bersih & akademik",
    swatches: ["#f8f9ff", "#eff4ff", "#00236f", "#fe932c"],
  },
  {
    id: "salt",
    name: "Salt",
    description: "Hangat & minimal",
    swatches: ["#f7f5f0", "#efede6", "#49453b", "#927957"],
  },
  {
    id: "matcha",
    name: "Matcha",
    description: "Alami & tenang",
    swatches: ["#f3f5ee", "#e8ede2", "#35533d", "#68866a"],
  },
  {
    id: "mint",
    name: "Mint",
    description: "Segar & ringan",
    swatches: ["#f2f8f6", "#e6f1ee", "#24594f", "#5b9284"],
  },
];

function isBrandTheme(value: string | null | undefined): value is BrandTheme {
  return BRAND_THEMES.some((theme) => theme === value);
}

function readTheme(): BrandTheme {
  const theme = document.documentElement.dataset.theme;
  return isBrandTheme(theme) ? theme : "nawasena";
}

function subscribe(callback: () => void) {
  const syncFromStorage = (event: StorageEvent) => {
    if (event.key !== THEME_STORAGE_KEY && event.key !== LEGACY_THEME_STORAGE_KEY) return;
    let saved: string | null = null;
    try {
      saved =
        window.localStorage.getItem(THEME_STORAGE_KEY) ??
        window.localStorage.getItem(LEGACY_THEME_STORAGE_KEY);
    } catch {
      saved = null;
    }
    document.documentElement.dataset.theme = isBrandTheme(saved) ? saved : "nawasena";
    callback();
  };
  window.addEventListener(THEME_EVENT, callback);
  window.addEventListener("storage", syncFromStorage);
  return () => {
    window.removeEventListener(THEME_EVENT, callback);
    window.removeEventListener("storage", syncFromStorage);
  };
}

function saveTheme(theme: BrandTheme) {
  document.documentElement.dataset.theme = theme;
  try {
    window.localStorage.setItem(THEME_STORAGE_KEY, theme);
  } catch {
    // The current tab still changes theme if browser storage is unavailable.
  }
  window.dispatchEvent(new Event(THEME_EVENT));
}

export function ThemeSwitcher({ className }: { className?: string }) {
  const theme = React.useSyncExternalStore(subscribe, readTheme, () => "nawasena");
  const [open, setOpen] = React.useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button
          type="button"
          variant="outline"
          aria-label="Ubah tema"
          title="Ubah tema"
          className={className}
        >
          <Palette aria-hidden="true" className="size-4" />
          <span>Tema</span>
        </Button>
      </DialogTrigger>
      <DialogContent className="w-[min(92vw,38rem)] p-5 sm:p-6">
        <DialogTitle className="pr-10 font-heading text-xl font-bold text-primary">Pilih Tema</DialogTitle>
        <DialogDescription className="mt-1 text-sm text-muted">
          Sesuaikan tampilan portal. Perubahan langsung tersimpan di perangkat ini.
        </DialogDescription>
        <fieldset className="mt-5 grid grid-cols-1 gap-3 min-[420px]:grid-cols-2">
          <legend className="sr-only">Pilihan tema</legend>
          {themes.map((item) => {
            const selected = theme === item.id;
            return (
              <button
                key={item.id}
                type="button"
                aria-pressed={selected}
                onClick={() => saveTheme(item.id)}
                className={`min-h-28 rounded-xl border p-3 text-left transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-secondary ${selected ? "border-primary bg-surface-muted ring-2 ring-primary/30" : "border-border bg-surface hover:bg-surface-muted"}`}
              >
                <span className="flex items-center justify-between">
                  <span className="flex gap-1.5" aria-hidden="true">
                    {item.swatches.map((color) => (
                      <span
                        key={color}
                        className="size-4 rounded-full border border-black/10"
                        style={{ backgroundColor: color }}
                      />
                    ))}
                  </span>
                  {selected ? <Check aria-label="Tema terpilih" className="size-4 text-primary" /> : null}
                </span>
                <span className="mt-3 block font-heading text-sm font-bold text-primary">{item.name}</span>
                <span className="mt-0.5 block text-xs text-muted">{item.description}</span>
              </button>
            );
          })}
        </fieldset>
      </DialogContent>
    </Dialog>
  );
}

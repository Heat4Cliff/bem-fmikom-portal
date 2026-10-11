import type { Metadata } from "next";
import { LandingSettingsAdmin } from "@/features/system/settings-admin";

export const metadata: Metadata = { title: "Pengaturan Landing Page" };
export default function SettingsPage() {
  return <LandingSettingsAdmin />;
}

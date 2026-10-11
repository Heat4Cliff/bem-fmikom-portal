import type { Metadata } from "next";
import { ActivityAdmin } from "@/features/system/activity-admin";

export const metadata: Metadata = { title: "Activity Log" };
export default function ActivityPage() {
  return <ActivityAdmin />;
}

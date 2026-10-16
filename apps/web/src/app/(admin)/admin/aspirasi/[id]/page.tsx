"use client";

import { useParams } from "next/navigation";
import { AspirationDetail } from "@/features/aspirations/aspiration-admin";

export default function AdminAspirationDetailPage() {
  const params = useParams<{ id: string }>();
  const id = Number(params.id);
  if (!Number.isSafeInteger(id) || id < 1) {
    return (
      <p role="alert" className="m-6 rounded-md bg-danger/5 p-4 text-sm text-danger-foreground">
        ID aspirasi tidak valid.
      </p>
    );
  }
  return <AspirationDetail id={id} />;
}

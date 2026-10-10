import type { Metadata } from "next";
import { DocumentArchive } from "@/features/documents/document-archive";

export const metadata: Metadata = { title: "Arsip Dokumen Internal" };

export default function DocumentArchivePage() {
  return <DocumentArchive />;
}

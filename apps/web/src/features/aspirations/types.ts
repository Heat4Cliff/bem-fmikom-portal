export type AspirationStatus = "PENDING" | "READ" | "RESOLVED";

export type AspirationDivision = {
  id: number;
  name: string;
  description?: string | null;
};

export type AspirationReceipt = {
  id: number;
  status: "PENDING";
  submitted_at: string;
};

export type Aspiration = {
  id: number;
  division_id: number;
  division_name: string;
  name: string | null;
  nim: string | null;
  email: string | null;
  subject: string;
  message: string;
  is_anonymous: boolean;
  status: AspirationStatus;
  submitted_at: string;
  read_at: string | null;
  resolved_at: string | null;
  created_at: string;
  updated_at: string;
};

export type AspirationPage = {
  items: Aspiration[];
  meta: { page: number; limit: number; total: number; total_pages: number };
};

export const aspirationStatusLabel: Record<AspirationStatus, string> = {
  PENDING: "Menunggu",
  READ: "Dibaca",
  RESOLVED: "Selesai",
};

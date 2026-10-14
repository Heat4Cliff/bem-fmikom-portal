export type PeriodStatus = "ACTIVE" | "ARCHIVED";
export type MemberStatus = "ACTIVE" | "INACTIVE";
export type MemberPosition =
  | "KETUA"
  | "WAKIL_KETUA"
  | "SEKRETARIS"
  | "BENDAHARA"
  | "KEPALA_BIDANG"
  | "ANGGOTA";

export type Period = {
  id: number;
  name: string;
  start_date: string;
  end_date: string;
  status: PeriodStatus;
  division_count: number;
  member_count: number;
};

export type Division = {
  id: number;
  period_id: number;
  period_name?: string;
  name: string;
  description?: string | null;
  status: "ACTIVE" | "INACTIVE";
  display_order: number;
  show_in_public_structure: boolean;
  head?: { id: number; name: string; position: string } | null;
  member_count: number;
};

export type OrganizationMember = {
  id: number;
  period_id: number;
  period_name?: string;
  division_id?: number | null;
  division_name?: string | null;
  name: string;
  nim?: string | null;
  position: MemberPosition;
  photo_url?: string | null;
  display_order: number;
  status: MemberStatus;
};

export type PageMeta = { page: number; limit: number; total: number; total_pages: number };
export type ApiPage<T> = { items: T[]; meta: PageMeta };

export const POSITION_OPTIONS: { value: MemberPosition; label: string }[] = [
  { value: "KETUA", label: "Ketua" },
  { value: "WAKIL_KETUA", label: "Wakil Ketua" },
  { value: "SEKRETARIS", label: "Sekretaris" },
  { value: "BENDAHARA", label: "Bendahara" },
  { value: "KEPALA_BIDANG", label: "Kepala Bidang" },
  { value: "ANGGOTA", label: "Anggota" },
];

export function positionLabel(value: string) {
  return POSITION_OPTIONS.find((option) => option.value === value)?.label ?? value;
}

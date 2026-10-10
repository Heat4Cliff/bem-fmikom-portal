export type RegistrationStatus = "DRAFT" | "OPEN" | "CLOSED";
export type PublicRegistrationStatus = "UPCOMING" | "OPEN" | "CLOSED";
export type RegistrationFieldType =
  | "TEXT"
  | "TEXTAREA"
  | "EMAIL"
  | "NUMBER"
  | "SELECT"
  | "RADIO"
  | "CHECKBOX"
  | "DATE";

export type RegistrationOption = { value: string; label: string; display_order?: number };
export type RegistrationField = {
  id?: number;
  field_key: string;
  label: string;
  field_type: RegistrationFieldType;
  placeholder?: string | null;
  required: boolean;
  display_order?: number;
  options: RegistrationOption[];
};
export type RegistrationForm = {
  id: number;
  period_id: number;
  period_name?: string;
  division_id: number | null;
  division_name: string | null;
  title: string;
  slug: string;
  description: string | null;
  open_at: string;
  close_at: string;
  status: RegistrationStatus;
  public_status?: PublicRegistrationStatus;
  submission_count: number;
};
export type PublicRegistration = Omit<
  RegistrationForm,
  "period_id" | "division_id" | "submission_count" | "status"
> & {
  status: PublicRegistrationStatus;
  period_name: string;
  fields: RegistrationField[];
};
export type PublicRegistrationPreview = {
  title: string;
  slug: string;
  description: string | null;
  period_name: string;
  division_name: string | null;
  open_at: string;
  close_at: string;
  public_status: PublicRegistrationStatus;
};
export type RegistrationSubmission = {
  id: number;
  registration_form_id: number;
  registration_title: string;
  division_name: string | null;
  full_name: string;
  nim: string;
  email: string;
  phone: string;
  submitted_at: string;
  answers?: {
    field_key: string;
    label: string;
    type: RegistrationFieldType;
    value_text: string | null;
    value_json?: unknown;
    display_order: number;
  }[];
};
export type PageMeta = { page: number; limit: number; total: number; total_pages: number };

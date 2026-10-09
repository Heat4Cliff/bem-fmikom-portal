"use client";

import { useParams } from "next/navigation";
import { MemberEditor } from "@/features/organization/organization-admin";

export default function EditMemberPage() {
  const { id } = useParams<{ id: string }>();
  const memberId = Number(id);
  return <MemberEditor memberId={Number.isSafeInteger(memberId) && memberId > 0 ? memberId : undefined} />;
}

import { Globe2, Link as LinkIcon } from "lucide-react";
import {
  SiFacebook,
  SiInstagram,
  SiTiktok,
  SiWhatsapp,
  SiX,
  SiYoutube,
} from "@icons-pack/react-simple-icons";

const iconByPlatform = {
  instagram: SiInstagram,
  tiktok: SiTiktok,
  youtube: SiYoutube,
  facebook: SiFacebook,
  x: SiX,
  twitter: SiX,
  whatsapp: SiWhatsapp,
  website: Globe2,
} as const;

export function SocialBrandIcon({ platform, className }: { platform: string; className?: string }) {
  const normalized = platform.trim().toLowerCase();
  if (normalized === "linkedin") {
    return (
      <span
        aria-hidden="true"
        className={`inline-grid size-4 place-items-center rounded-[2px] bg-current font-sans text-[10px] font-black leading-none text-surface ${className ?? ""}`}
      >
        in
      </span>
    );
  }
  const Icon = iconByPlatform[normalized as keyof typeof iconByPlatform] ?? LinkIcon;
  return <Icon aria-hidden="true" className={className ?? "size-4"} />;
}

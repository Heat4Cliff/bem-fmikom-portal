import type { PublicLandingSettings } from "@/features/content/api";

export function landingSocialLinks(settings: PublicLandingSettings) {
  const configured = (settings.social_links ?? [])
    .filter((item) => item.active && item.url.trim())
    .sort((a, b) => a.order - b.order)
    .map((item) => ({ platform: item.platform, label: item.label, href: item.url }));
  if (configured.length) return configured;

  const legacy = [
    { platform: "instagram", label: "Instagram", href: settings.instagram_url },
    { platform: "youtube", label: "YouTube", href: settings.youtube_url },
    { platform: "tiktok", label: "TikTok", href: settings.tiktok_url },
    { platform: "website", label: "Situs lainnya", href: settings.other_social_url },
  ];
  return legacy.flatMap(({ platform, label, href }) => (href ? [{ platform, label, href }] : []));
}

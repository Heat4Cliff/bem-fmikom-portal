import { describe, expect, it } from "vitest";
import type { PublicLandingSettings } from "@/features/content/api";
import { landingSocialLinks } from "./landing-social-links";

const base = {
  social_links: [],
  instagram_url: null,
  youtube_url: null,
  tiktok_url: null,
  other_social_url: null,
} as unknown as PublicLandingSettings;

describe("public landing social links", () => {
  it("orders active configured links and omits disabled entries", () => {
    const links = landingSocialLinks({
      ...base,
      social_links: [
        { platform: "tiktok", label: "TikTok", url: "https://tiktok.com/@bem", order: 2, active: true },
        {
          platform: "instagram",
          label: "Instagram",
          url: "https://instagram.com/bem",
          order: 1,
          active: true,
        },
        { platform: "youtube", label: "YouTube", url: "https://youtube.com/@bem", order: 3, active: false },
      ],
    });
    expect(links.map(({ platform }) => platform)).toEqual(["instagram", "tiktok"]);
  });

  it("keeps legacy social settings visible for installations without the new model", () => {
    expect(landingSocialLinks({ ...base, instagram_url: "https://instagram.com/bem" })).toEqual([
      { platform: "instagram", label: "Instagram", href: "https://instagram.com/bem" },
    ]);
  });
});

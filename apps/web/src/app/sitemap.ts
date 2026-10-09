import type { MetadataRoute } from "next";
import { getPublicSiteOrigin } from "@/lib/seo/site-url";

export default function sitemap(): MetadataRoute.Sitemap {
  const siteOrigin = getPublicSiteOrigin();
  if (!siteOrigin) return [];

  return [
    {
      url: new URL("/", siteOrigin).toString(),
      changeFrequency: "weekly",
      priority: 1,
    },
  ];
}

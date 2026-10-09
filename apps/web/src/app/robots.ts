import type { MetadataRoute } from "next";
import { getPublicSiteOrigin } from "@/lib/seo/site-url";

export default function robots(): MetadataRoute.Robots {
  const siteOrigin = getPublicSiteOrigin();
  const sitemap = siteOrigin ? `${siteOrigin}/sitemap.xml` : undefined;

  return {
    rules: { userAgent: "*", allow: "/", disallow: "/admin/" },
    ...(sitemap ? { sitemap } : {}),
  };
}

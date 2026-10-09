export function getPublicSiteOrigin(): string | undefined {
  const configuredSiteUrl = process.env.NEXT_PUBLIC_SITE_URL;
  if (!configuredSiteUrl) return undefined;

  try {
    const parsedUrl = new URL(configuredSiteUrl);
    if (parsedUrl.protocol !== "https:" && parsedUrl.protocol !== "http:") return undefined;
    if (parsedUrl.username || parsedUrl.password) return undefined;
    return parsedUrl.origin;
  } catch {
    return undefined;
  }
}

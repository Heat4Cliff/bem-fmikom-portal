const defaultApiBaseURL = "http://localhost:8080/api/v1";
const internalMediaPath = /^\/api\/v1\/public\/media\/([1-9]\d*)\/?$/;

function getApiOrigin() {
  const apiBaseURL = process.env.NEXT_PUBLIC_API_BASE_URL ?? defaultApiBaseURL;
  try {
    return new URL(apiBaseURL).origin;
  } catch {
    return new URL(defaultApiBaseURL).origin;
  }
}

function internalMediaURL(id: string) {
  return new URL(`/api/v1/public/media/${id}`, getApiOrigin()).toString();
}

export function resolveMediaUrl(value: string | null | undefined, allowBlobPreview = false): string | null {
  const candidate = value?.trim();
  if (!candidate) return null;

  const mediaReference = /^media:([1-9]\d*)$/i.exec(candidate);
  if (mediaReference) return internalMediaURL(mediaReference[1]);

  if (allowBlobPreview && candidate.startsWith("blob:")) {
    try {
      return new URL(candidate).protocol === "blob:" ? candidate : null;
    } catch {
      return null;
    }
  }

  try {
    const parsed = new URL(candidate);
    const mediaPath = internalMediaPath.exec(parsed.pathname);
    if (mediaPath) return internalMediaURL(mediaPath[1]);
    return parsed.protocol === "https:" ? parsed.toString() : null;
  } catch {
    if (candidate.startsWith("/")) {
      const mediaPath = internalMediaPath.exec(candidate);
      if (mediaPath) return internalMediaURL(mediaPath[1]);
    }
    return null;
  }
}

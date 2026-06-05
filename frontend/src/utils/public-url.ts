function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, '');
}

function isLocalHost(hostname: string): boolean {
  return hostname === 'localhost' || hostname === '127.0.0.1' || hostname === '::1';
}

function configuredPublicBaseURL(): string {
  return trimTrailingSlash(import.meta.env.VITE_PUBLIC_BASE_URL || '');
}

export function resolvePublicBaseURL(): string {
  const configured = configuredPublicBaseURL();
  if (configured) return configured;

  return window.location.origin;
}

export function resolveShareURL(url: string, token?: string): string {
  const baseURL = resolvePublicBaseURL();
  const path = token ? `/s/${token}` : extractPath(url);
  return `${baseURL}${path}`;
}

export function resolveAPIBaseURL(rawBaseURL: string | undefined): string {
  const fallback = '/';
  const baseURL = rawBaseURL || fallback;

  if (!baseURL.startsWith('http')) return baseURL;

  try {
    const parsed = new URL(baseURL);
    if (!isLocalHost(parsed.hostname) || isLocalHost(window.location.hostname)) {
      return baseURL;
    }

    parsed.hostname = window.location.hostname;
    return parsed.toString().replace(/\/$/, '');
  } catch {
    return baseURL;
  }
}

function extractPath(url: string): string {
  try {
    return new URL(url).pathname;
  } catch {
    return url.startsWith('/') ? url : `/s/${url}`;
  }
}

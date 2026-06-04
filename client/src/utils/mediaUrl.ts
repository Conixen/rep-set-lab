const apiBase = import.meta.env.VITE_API_BASE_URL ?? ''

export function mediaUrl(url: string | null | undefined): string | null {
  if (!url) return null
  return url.startsWith('/') ? apiBase + url : url
}

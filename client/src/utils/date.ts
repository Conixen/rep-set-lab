// formatDateTime: "28 May, 14:05" — used in admin request log and history
export function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleDateString('sv-SE', {
    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
  })
}

// formatDateLong: "Mon, 28 May, 14:05" — used in workout history list
export function formatDateLong(iso: string): string {
  return new Date(iso).toLocaleDateString('sv-SE', {
    weekday: 'short', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
  })
}

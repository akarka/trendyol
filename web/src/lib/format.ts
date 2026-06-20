export function formatDate(iso: string): string {
  const d = new Date(iso)
  if (isNaN(d.getTime())) return iso
  return d.toLocaleString('tr-TR', { dateStyle: 'short', timeStyle: 'short' })
}

export function formatTL(n: number): string {
  return n.toLocaleString('tr-TR', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) + ' TL'
}

export function customerName(addr?: { firstName?: string; lastName?: string }): string {
  if (!addr) return '—'
  return [addr.firstName, addr.lastName].filter(Boolean).join(' ') || '—'
}

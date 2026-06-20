const STATUS_STYLES: Record<string, string> = {
  Created: 'bg-green-100 text-green-800',
  Cancelled: 'bg-red-100 text-red-800',
  Delivered: 'bg-blue-100 text-blue-800',
  UnSupplied: 'bg-yellow-100 text-yellow-800',
}

export function Badge({ status }: { status: string }) {
  const style = STATUS_STYLES[status] ?? 'bg-gray-100 text-gray-700'
  return (
    <span className={`inline-block rounded px-2 py-0.5 text-xs font-medium ${style}`}>{status}</span>
  )
}

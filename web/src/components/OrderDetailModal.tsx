import { OrderRow, printOrder } from '../api/orders'
import { Badge } from './Badge'
import { Spinner } from './Spinner'
import { customerName, formatDate, formatTL } from '../lib/format'
import { useToast } from '../context/ToastContext'
import { useMutation, useQueryClient } from '@tanstack/react-query'

export function OrderDetailModal({ order, onClose }: { order: OrderRow; onClose: () => void }) {
  const toast = useToast()
  const qc = useQueryClient()
  const p = order.payload
  const lines = p.lines ?? []
  const total = lines.reduce((sum, l) => sum + l.amount * l.quantity, 0)
  const addr = p.shipmentAddress

  const reprint = useMutation({
    mutationFn: () => printOrder(order.order_id),
    onSuccess: () => {
      toast.show('success', 'Baskı kuyruğa alındı')
      qc.invalidateQueries({ queryKey: ['printer-status'] })
    },
    onError: (e: any) => toast.show('error', e?.response?.data?.error ?? 'Baskı başarısız'),
  })

  return (
    <div
      className="fixed inset-0 z-40 flex items-end justify-center bg-black/40 sm:items-center"
      onClick={onClose}
    >
      <div
        className="max-h-[90vh] w-full overflow-y-auto bg-white p-5 sm:max-w-lg sm:rounded-lg"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-start justify-between">
          <div>
            <div className="text-lg font-semibold">{order.order_number}</div>
            <div className="text-sm text-gray-500">{formatDate(order.created_at)}</div>
          </div>
          <Badge status={order.package_status} />
        </div>

        <dl className="mt-4 grid grid-cols-2 gap-2 text-sm">
          <div>
            <dt className="text-gray-500">Müşteri</dt>
            <dd>{customerName(addr)}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Kargo</dt>
            <dd>{p.cargoProviderName ?? '—'}</dd>
          </div>
        </dl>

        <div className="mt-4">
          <div className="mb-1 text-sm font-medium text-gray-700">Ürünler</div>
          <div className="divide-y rounded border">
            {lines.map((l, i) => (
              <div key={i} className="flex items-center justify-between px-3 py-2 text-sm">
                <span>
                  {l.quantity}× {l.productName}
                </span>
                <span className="tabular-nums">{formatTL(l.amount * l.quantity)}</span>
              </div>
            ))}
            <div className="flex items-center justify-between px-3 py-2 text-sm font-semibold">
              <span>Toplam</span>
              <span className="tabular-nums">{formatTL(total)}</span>
            </div>
          </div>
        </div>

        {addr && (
          <div className="mt-4 text-sm">
            <div className="text-gray-500">Adres</div>
            <div>
              {[addr.address1, addr.district, addr.city, addr.postalCode]
                .filter(Boolean)
                .join(', ')}
            </div>
          </div>
        )}

        <div className="mt-6 flex gap-2">
          <button
            onClick={onClose}
            className="h-10 flex-1 rounded border border-gray-300 text-sm font-medium"
          >
            Kapat
          </button>
          <button
            onClick={() => reprint.mutate()}
            disabled={reprint.isPending}
            className="flex h-10 flex-1 items-center justify-center rounded bg-gray-900 text-sm font-medium text-white disabled:opacity-60"
          >
            {reprint.isPending ? <Spinner /> : 'Yeniden Yazdır'}
          </button>
        </div>
      </div>
    </div>
  )
}

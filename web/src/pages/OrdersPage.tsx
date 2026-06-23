import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { exportOrders, getOrders, OrderRow, printOrder } from '../api/orders'
import { Badge } from '../components/Badge'
import { Spinner } from '../components/Spinner'
import { OrderDetailModal } from '../components/OrderDetailModal'
import { useToast } from '../context/ToastContext'
import { customerName, formatDate } from '../lib/format'

const STATUSES = ['', 'Created', 'Cancelled', 'Delivered', 'UnSupplied']
const STATUS_LABEL: Record<string, string> = {
  '': 'Tümü',
  Created: 'Oluşturuldu',
  Cancelled: 'İptal',
  Delivered: 'Teslim',
  UnSupplied: 'Tedarik Yok',
}
const LIMIT = 50

export function OrdersPage() {
  const [status, setStatus] = useState('')
  const [page, setPage] = useState(0)
  const [selected, setSelected] = useState<OrderRow | null>(null)
  const toast = useToast()
  const qc = useQueryClient()

  const { data, isLoading, isError } = useQuery({
    queryKey: ['orders', { status, page }],
    queryFn: () => getOrders({ status, limit: LIMIT, offset: page * LIMIT }),
  })

  const print = useMutation({
    mutationFn: (orderID: string) => printOrder(orderID),
    onSuccess: () => {
      toast.show('success', 'Baskı kuyruğa alındı')
      qc.invalidateQueries({ queryKey: ['printer-status'] })
    },
    onError: (e: any) => toast.show('error', e?.response?.data?.error ?? 'Baskı başarısız'),
  })

  const onFilter = (s: string) => {
    setStatus(s)
    setPage(0)
  }

  const exportMutation = useMutation({
    mutationFn: () => exportOrders(status),
    onError: () => toast.show('error', 'Excel oluşturulamadı'),
  })

  const orders = data ?? []

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-xl font-semibold">Siparişler</h1>
        <button
          onClick={() => exportMutation.mutate()}
          disabled={exportMutation.isPending}
          className="h-9 rounded border border-gray-300 bg-white px-3 text-sm font-medium text-gray-700 disabled:opacity-50"
        >
          {exportMutation.isPending ? '...' : 'Excel’e Aktar'}
        </button>
      </div>

      <div className="mb-4 flex flex-wrap gap-2">
        {STATUSES.map((s) => (
          <button
            key={s || 'all'}
            onClick={() => onFilter(s)}
            className={`rounded-full px-3 py-1 text-sm ${
              status === s ? 'bg-gray-900 text-white' : 'bg-white text-gray-700 hover:bg-gray-200'
            }`}
          >
            {STATUS_LABEL[s]}
          </button>
        ))}
      </div>

      {isLoading && (
        <div className="flex justify-center py-10 text-gray-400">
          <Spinner />
        </div>
      )}
      {isError && <div className="py-10 text-center text-red-600">Bağlantı hatası</div>}
      {!isLoading && !isError && orders.length === 0 && (
        <div className="py-10 text-center text-gray-400">Henüz sipariş yok</div>
      )}

      {orders.length > 0 && (
        <>
          {/* Desktop tablo */}
          <table className="hidden w-full overflow-hidden rounded-lg bg-white text-sm shadow sm:table">
            <thead className="bg-gray-50 text-left text-gray-500">
              <tr>
                <th className="px-4 py-2">Sipariş No</th>
                <th className="px-4 py-2">Müşteri</th>
                <th className="px-4 py-2">Durum</th>
                <th className="px-4 py-2">Tarih</th>
                <th className="px-4 py-2"></th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {orders.map((o) => (
                <tr key={o.uuid} className="cursor-pointer hover:bg-gray-50" onClick={() => setSelected(o)}>
                  <td className="px-4 py-2 font-medium">{o.order_number}</td>
                  <td className="px-4 py-2">{customerName(o.payload.shipmentAddress)}</td>
                  <td className="px-4 py-2">
                    <Badge status={o.package_status} />
                  </td>
                  <td className="px-4 py-2 text-gray-500">{formatDate(o.created_at)}</td>
                  <td className="px-4 py-2 text-right">
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        print.mutate(o.order_id)
                      }}
                      disabled={print.isPending && print.variables === o.order_id}
                      className="h-8 rounded bg-gray-900 px-3 text-xs font-medium text-white disabled:opacity-60"
                    >
                      {print.isPending && print.variables === o.order_id ? '...' : 'Yazdır'}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {/* Mobil kartlar */}
          <div className="grid gap-2 sm:hidden">
            {orders.map((o) => (
              <div key={o.uuid} className="rounded-lg bg-white p-3 shadow" onClick={() => setSelected(o)}>
                <div className="flex items-center justify-between">
                  <span className="font-medium">{o.order_number}</span>
                  <Badge status={o.package_status} />
                </div>
                <div className="mt-1 text-sm text-gray-600">{customerName(o.payload.shipmentAddress)}</div>
                <div className="mt-1 flex items-center justify-between">
                  <span className="text-xs text-gray-400">{formatDate(o.created_at)}</span>
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      print.mutate(o.order_id)
                    }}
                    disabled={print.isPending && print.variables === o.order_id}
                    className="h-10 rounded bg-gray-900 px-4 text-sm font-medium text-white disabled:opacity-60"
                  >
                    {print.isPending && print.variables === o.order_id ? '...' : 'Yazdır'}
                  </button>
                </div>
              </div>
            ))}
          </div>

          <div className="mt-4 flex items-center justify-between">
            <button
              onClick={() => setPage((p) => Math.max(0, p - 1))}
              disabled={page === 0}
              className="h-10 rounded border border-gray-300 px-4 text-sm disabled:opacity-40"
            >
              Önceki
            </button>
            <span className="text-sm text-gray-500">Sayfa {page + 1}</span>
            <button
              onClick={() => setPage((p) => p + 1)}
              disabled={orders.length < LIMIT}
              className="h-10 rounded border border-gray-300 px-4 text-sm disabled:opacity-40"
            >
              Sonraki
            </button>
          </div>
        </>
      )}

      {selected && <OrderDetailModal order={selected} onClose={() => setSelected(null)} />}
    </div>
  )
}

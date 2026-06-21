import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { getProducts, Product } from '../api/products'
import { createManualOrder } from '../api/orders'
import { getSettings } from '../api/settings'
import { buildLabelLines, labelDataUrl, printLabel } from '../lib/label'
import { cellCount, parseLayout } from '../lib/labelLayout'
import { Spinner } from '../components/Spinner'
import { SheetPreview } from '../components/SheetPreview'
import { useToast } from '../context/ToastContext'

interface Row {
  sku: string
  quantity: number
}

export function ManualOrderPage() {
  const [customer, setCustomer] = useState('')
  const [rows, setRows] = useState<Row[]>([{ sku: '', quantity: 1 }])
  const [activeOnly, setActiveOnly] = useState(true)
  const [cell, setCell] = useState(0)
  const toast = useToast()
  const qc = useQueryClient()

  const { data: products, isLoading } = useQuery({
    queryKey: ['products', { activeOnly }],
    queryFn: () => getProducts(activeOnly),
  })

  const { data: settings } = useQuery({ queryKey: ['settings'], queryFn: getSettings })
  const layout = useMemo(() => parseLayout(settings), [settings])
  const cells = cellCount(layout)

  const bySku = useMemo(() => {
    const m = new Map<string, Product>()
    for (const p of products ?? []) m.set(p.sku, p)
    return m
  }, [products])

  const items = useMemo(
    () =>
      rows
        .filter((r) => r.sku && bySku.has(r.sku))
        .map((r) => {
          const p = bySku.get(r.sku)!
          return { quantity: r.quantity, name: p.name, price: p.price }
        }),
    [rows, bySku],
  )

  const previewLines = useMemo(
    () => buildLabelLines({ orderNumber: 'ÖNİZLEME', customer, items }),
    [customer, items],
  )

  const previewUrl = useMemo(
    () => (items.length > 0 ? labelDataUrl(previewLines) : ''),
    [items.length, previewLines],
  )

  const save = useMutation({
    mutationFn: () =>
      createManualOrder(
        customer,
        rows.filter((r) => r.sku).map((r) => ({ sku: r.sku, quantity: r.quantity })),
      ),
    onSuccess: (res) => {
      toast.show('success', `Sipariş kaydedildi: ${res.order_number}`)
      const printCell = cells > 0 ? Math.min(cell, cells - 1) : 0
      printLabel(buildLabelLines({ orderNumber: res.order_number, customer, items }), layout, printCell)
      qc.invalidateQueries({ queryKey: ['orders'] })
      qc.invalidateQueries({ queryKey: ['printer-status'] })
      setCustomer('')
      setRows([{ sku: '', quantity: 1 }])
      if (cells > 0) setCell((c) => (c + 1) % cells) // sıradaki sticker hücresine ilerle
    },
    onError: (e: any) => toast.show('error', e?.response?.data?.error ?? 'Kayıt başarısız'),
  })

  const setRow = (i: number, patch: Partial<Row>) =>
    setRows((rs) => rs.map((r, idx) => (idx === i ? { ...r, ...patch } : r)))
  const addRow = () => setRows((rs) => [...rs, { sku: '', quantity: 1 }])
  const removeRow = (i: number) => setRows((rs) => rs.filter((_, idx) => idx !== i))

  const canSave = items.length > 0 && !save.isPending

  return (
    <div className="max-w-3xl">
      <h1 className="mb-4 text-xl font-semibold">Manuel Sipariş</h1>

      {isLoading ? (
        <div className="flex justify-center py-10 text-gray-400">
          <Spinner />
        </div>
      ) : (
        <div className="grid gap-6 lg:grid-cols-2">
          <div className="space-y-4">
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">Müşteri Adı</label>
              <input
                value={customer}
                onChange={(e) => setCustomer(e.target.value)}
                placeholder="Ad Soyad"
                className="h-10 w-full rounded border border-gray-300 px-3 text-sm"
              />
            </div>

            <label className="flex items-center gap-2 text-sm text-gray-600">
              <input
                type="checkbox"
                checked={activeOnly}
                onChange={(e) => setActiveOnly(e.target.checked)}
              />
              Sadece aktif ürünler
            </label>

            <div className="space-y-2">
              <span className="block text-sm font-medium text-gray-700">Ürünler</span>
              {rows.map((row, i) => (
                <div key={i} className="flex gap-2">
                  <select
                    value={row.sku}
                    onChange={(e) => setRow(i, { sku: e.target.value })}
                    className="h-10 min-w-0 flex-1 rounded border border-gray-300 px-2 text-sm"
                  >
                    <option value="">Ürün seç…</option>
                    {(products ?? []).map((p) => (
                      <option key={p.sku} value={p.sku}>
                        {p.name} — {p.price.toFixed(2)} TL ({p.category})
                      </option>
                    ))}
                  </select>
                  <input
                    type="number"
                    min={1}
                    value={row.quantity}
                    onChange={(e) => setRow(i, { quantity: Math.max(1, Number(e.target.value) || 1) })}
                    className="h-10 w-16 rounded border border-gray-300 px-2 text-sm"
                  />
                  <button
                    onClick={() => removeRow(i)}
                    disabled={rows.length === 1}
                    className="h-10 w-10 shrink-0 rounded border border-gray-300 text-gray-500 disabled:opacity-40"
                    title="Satırı sil"
                  >
                    ×
                  </button>
                </div>
              ))}
              <button
                onClick={addRow}
                className="h-9 rounded border border-dashed border-gray-300 px-3 text-sm text-gray-600 hover:bg-gray-50"
              >
                + Ürün ekle
              </button>
            </div>

            <button
              onClick={() => save.mutate()}
              disabled={!canSave}
              className="h-11 w-full rounded bg-gray-900 px-4 text-sm font-medium text-white disabled:opacity-50"
            >
              {save.isPending ? 'Kaydediliyor…' : 'Kaydet ve Yazdır'}
            </button>
          </div>

          <div className="space-y-4">
            <div>
              <span className="mb-1 block text-sm font-medium text-gray-700">Etiket Önizleme</span>
              <div className="flex min-h-[160px] items-center justify-center rounded border bg-gray-50 p-2">
                {previewUrl ? (
                  <img src={previewUrl} alt="etiket önizleme" className="max-w-full" />
                ) : (
                  <span className="text-sm text-gray-400">Ürün seçince önizleme görünür</span>
                )}
              </div>
            </div>

            <div>
              <span className="mb-1 block text-sm font-medium text-gray-700">
                Yazdırılacak hücre ({cells > 0 ? cell + 1 : 0}/{cells})
              </span>
              <p className="mb-2 text-xs text-gray-400">
                Şablonlu sticker sayfasında etiketin basılacağı konumu seçin. Yerleşim Ayarlar'dan
                düzenlenir.
              </p>
              <SheetPreview
                layout={layout}
                selected={cells > 0 ? Math.min(cell, cells - 1) : undefined}
                onSelect={setCell}
              />
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

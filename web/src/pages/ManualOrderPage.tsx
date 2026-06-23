import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { getProducts, Product } from '../api/products'
import { createManualOrder } from '../api/orders'
import { getSettings, updateSetting } from '../api/settings'
import { buildLabelLines, labelDataUrl, printLabel, downloadLabelBmp } from '../lib/label'
import { cellCount, LABEL_LAYOUT_KEY, LabelLayout, parseLayout } from '../lib/labelLayout'
import { Spinner } from '../components/Spinner'
import { SheetPreview } from '../components/SheetPreview'
import { LabelPaddingBox } from '../components/LabelPaddingBox'
import { useToast } from '../context/ToastContext'

// Sayfa eni/boyu ve sütun/satır sayısını düzenler. Etiket eni/boyu bunlardan
// int(sayfa/sütun) ve int(sayfa/satır) olarak türetilir (bkz. lib/labelLayout.normalizeLayout) —
// böylece sayfada hiçbir zaman boş hücre/şerit kalmaz.
function PaperSizeFields({
  pageWidthMm,
  pageHeightMm,
  onChange,
}: {
  pageWidthMm: number
  pageHeightMm: number
  onChange: (patch: { pageWidthMm?: number; pageHeightMm?: number }) => void
}) {
  return (
    <div className="flex items-end gap-2 text-xs text-gray-600">
      <label>
        Sayfa En (mm)
        <input
          type="number"
          step="any"
          min={1}
          value={pageWidthMm}
          onChange={(e) => onChange({ pageWidthMm: Number(e.target.value) || 0 })}
          className="mt-1 block h-8 w-20 rounded border border-gray-300 px-2"
        />
      </label>
      <label>
        Sayfa Boy (mm)
        <input
          type="number"
          step="any"
          min={1}
          value={pageHeightMm}
          onChange={(e) => onChange({ pageHeightMm: Number(e.target.value) || 0 })}
          className="mt-1 block h-8 w-20 rounded border border-gray-300 px-2"
        />
      </label>
    </div>
  )
}

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

  // Kağıt Ayarları: sayfa eni/boyu + sütun/satır. Etiket eni/boyu bunlardan türetilir.
  const [paper, setPaper] = useState<Pick<LabelLayout, 'pageWidthMm' | 'pageHeightMm' | 'columns' | 'rows'>>({
    pageWidthMm: layout.pageWidthMm,
    pageHeightMm: layout.pageHeightMm,
    columns: layout.columns,
    rows: layout.rows,
  })
  useEffect(() => {
    setPaper({
      pageWidthMm: layout.pageWidthMm,
      pageHeightMm: layout.pageHeightMm,
      columns: layout.columns,
      rows: layout.rows,
    })
  }, [layout.pageWidthMm, layout.pageHeightMm, layout.columns, layout.rows])

  const paperDirty =
    paper.pageWidthMm !== layout.pageWidthMm ||
    paper.pageHeightMm !== layout.pageHeightMm ||
    paper.columns !== layout.columns ||
    paper.rows !== layout.rows

  const savePaper = useMutation({
    mutationFn: () => updateSetting(LABEL_LAYOUT_KEY, JSON.stringify({ ...layout, ...paper })),
    onSuccess: () => {
      toast.show('success', 'Kağıt ayarları kaydedildi')
      qc.invalidateQueries({ queryKey: ['settings'] })
    },
    onError: () => toast.show('error', 'Kaydedilemedi'),
  })

  // Etiket Ayarları: hücre içi baskı alanı boşluğu (LabelPaddingBox ile sürüklenerek ayarlanır).
  const [padding, setPadding] = useState(layout.paddingMm)
  useEffect(() => setPadding(layout.paddingMm), [layout.paddingMm])
  const paddingDirty = padding !== layout.paddingMm

  const savePadding = useMutation({
    mutationFn: () => updateSetting(LABEL_LAYOUT_KEY, JSON.stringify({ ...layout, paddingMm: padding })),
    onSuccess: () => {
      toast.show('success', 'Etiket ayarları kaydedildi')
      qc.invalidateQueries({ queryKey: ['settings'] })
    },
    onError: () => toast.show('error', 'Kaydedilemedi'),
  })

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
                <div key={i} className="grid grid-cols-[minmax(0,1fr)_3.5rem_2.5rem] gap-2">
                  <select
                    value={row.sku}
                    onChange={(e) => setRow(i, { sku: e.target.value })}
                    className="h-10 w-full min-w-0 rounded border border-gray-300 px-2 text-sm"
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
                    className="h-10 w-full min-w-0 rounded border border-gray-300 px-2 text-sm"
                  />
                  <button
                    onClick={() => removeRow(i)}
                    disabled={rows.length === 1}
                    className="h-10 w-full rounded border border-gray-300 text-gray-500 disabled:opacity-40"
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
            <h2 className="text-base font-semibold text-gray-800">Etiket</h2>

            <div>
              <span className="mb-1 block text-sm font-medium text-gray-700">Etiket Ayarları</span>
              <p className="mb-2 text-xs text-gray-400">
                Kutu, sayfadan türetilen hücre ebadını ({layout.labelWidthMm}×{layout.labelHeightMm} mm)
                yansıtır. İçindeki kutuyu sağ-alt köşeden sürükleyerek baskı alanı (kenar boşluğu:{' '}
                {padding} mm) ayarlanır.
              </p>
              <LabelPaddingBox
                labelWidthMm={layout.labelWidthMm}
                labelHeightMm={layout.labelHeightMm}
                paddingMm={padding}
                onPaddingChange={setPadding}
                previewUrl={previewUrl}
              />
              <div className="mt-2 flex flex-wrap gap-2">
                <button
                  onClick={() => savePadding.mutate()}
                  disabled={!paddingDirty || savePadding.isPending}
                  className="h-9 rounded bg-gray-900 px-3 text-sm font-medium text-white disabled:opacity-40"
                >
                  {savePadding.isPending ? 'Kaydediliyor…' : 'Etiket Ayarlarını Kaydet'}
                </button>
                <button
                  onClick={() =>
                    downloadLabelBmp(
                      previewLines,
                      `etiket-${new Date().toISOString().slice(0, 19).replace(/[:T]/g, '')}`,
                    )
                  }
                  disabled={items.length === 0}
                  className="h-9 rounded border border-gray-300 px-3 text-sm text-gray-600 hover:bg-gray-50 disabled:opacity-40"
                >
                  .bmp indir
                </button>
              </div>
            </div>

            <div>
              <div className="mb-1 flex flex-wrap items-end justify-between gap-2">
                <span className="text-sm font-medium text-gray-700">
                  Kağıt Ayarları ({cells > 0 ? cell + 1 : 0}/{cells})
                </span>
                <PaperSizeFields
                  pageWidthMm={paper.pageWidthMm}
                  pageHeightMm={paper.pageHeightMm}
                  onChange={(patch) => setPaper((p) => ({ ...p, ...patch }))}
                />
              </div>
              <p className="mb-2 text-xs text-gray-400">
                Şablonlu sticker sayfasında etiketin basılacağı konumu seçin. Etiket ebadı, sayfa ile
                sütun/satır sayısından otomatik hesaplanır — sayfada boş alan kalmaz.
              </p>

              <div className="mb-3 flex flex-wrap items-end gap-2 text-xs text-gray-600">
                <label>
                  Sütun
                  <input
                    type="number"
                    min={1}
                    value={paper.columns}
                    onChange={(e) =>
                      setPaper((p) => ({ ...p, columns: Math.max(1, Number(e.target.value) || 1) }))
                    }
                    className="mt-1 block h-8 w-16 rounded border border-gray-300 px-2"
                  />
                </label>
                <label>
                  Satır
                  <input
                    type="number"
                    min={1}
                    value={paper.rows}
                    onChange={(e) =>
                      setPaper((p) => ({ ...p, rows: Math.max(1, Number(e.target.value) || 1) }))
                    }
                    className="mt-1 block h-8 w-16 rounded border border-gray-300 px-2"
                  />
                </label>
                <button
                  onClick={() => savePaper.mutate()}
                  disabled={!paperDirty || savePaper.isPending}
                  className="h-8 rounded bg-gray-900 px-3 text-xs font-medium text-white disabled:opacity-40"
                >
                  {savePaper.isPending ? 'Kaydediliyor…' : 'Kağıt Ayarlarını Kaydet'}
                </button>
              </div>

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

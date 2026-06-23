import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createProduct, getBrands, getCategories, Product, ProductInput, updateProduct } from '../api/products'
import { Spinner } from './Spinner'
import { useToast } from '../context/ToastContext'

function toInput(p?: Product): ProductInput {
  return {
    sku: p?.sku ?? '',
    barcode: p?.barcode ?? '',
    name: p?.name ?? '',
    marketplace_name: p?.marketplace_name ?? '',
    category: p?.category ?? '',
    brand: p?.brand ?? '',
    net_weight: p?.net_weight ?? null,
    unit: p?.unit ?? '',
    price: p?.price ?? 0,
    vat_rate: p?.vat_rate ?? null,
    is_active: p?.is_active ?? true,
    description: p?.description ?? '',
  }
}

export function ProductFormModal({ product, onClose }: { product?: Product; onClose: () => void }) {
  const isEdit = !!product
  const [form, setForm] = useState<ProductInput>(toInput(product))
  const toast = useToast()
  const qc = useQueryClient()

  const { data: categories } = useQuery({ queryKey: ['categories'], queryFn: getCategories })
  const { data: brands } = useQuery({ queryKey: ['brands'], queryFn: getBrands })

  const save = useMutation({
    mutationFn: () => (isEdit ? updateProduct(product!.sku, form) : createProduct(form)),
    onSuccess: () => {
      toast.show('success', isEdit ? 'Ürün güncellendi' : 'Ürün eklendi')
      qc.invalidateQueries({ queryKey: ['products'] })
      qc.invalidateQueries({ queryKey: ['categories'] })
      qc.invalidateQueries({ queryKey: ['brands'] })
      onClose()
    },
    onError: (e: any) => toast.show('error', e?.response?.data?.error ?? 'Kaydedilemedi'),
  })

  const set = <K extends keyof ProductInput>(k: K, v: ProductInput[K]) =>
    setForm((f) => ({ ...f, [k]: v }))

  return (
    <div
      className="fixed inset-0 z-40 flex items-end justify-center bg-black/40 sm:items-center"
      onClick={onClose}
    >
      <div
        className="max-h-[90vh] w-full overflow-y-auto bg-white p-5 sm:max-w-lg sm:rounded-lg"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="mb-4 text-lg font-semibold">{isEdit ? 'Ürünü Düzenle' : 'Yeni Ürün'}</h2>

        <div className="grid grid-cols-2 gap-3 text-sm">
          <label className="col-span-1">
            SKU
            <input
              value={form.sku}
              onChange={(e) => set('sku', e.target.value)}
              disabled={isEdit}
              className="mt-1 h-9 w-full rounded border border-gray-300 px-2 disabled:bg-gray-100"
            />
          </label>
          <label className="col-span-1">
            Barkod
            <input
              value={form.barcode}
              onChange={(e) => set('barcode', e.target.value)}
              className="mt-1 h-9 w-full rounded border border-gray-300 px-2"
            />
          </label>
          <label className="col-span-2">
            Ürün Adı
            <input
              value={form.name}
              onChange={(e) => set('name', e.target.value)}
              className="mt-1 h-9 w-full rounded border border-gray-300 px-2"
            />
          </label>
          <label className="col-span-2">
            Ty Ürün Adı
            <input
              value={form.marketplace_name}
              onChange={(e) => set('marketplace_name', e.target.value)}
              className="mt-1 h-9 w-full rounded border border-gray-300 px-2"
            />
          </label>
          <label className="col-span-1">
            Kategori
            <input
              value={form.category}
              onChange={(e) => set('category', e.target.value)}
              list="category-options"
              className="mt-1 h-9 w-full rounded border border-gray-300 px-2"
            />
            <datalist id="category-options">
              {categories?.map((c) => <option key={c} value={c} />)}
            </datalist>
          </label>
          <label className="col-span-1">
            Marka
            <input
              value={form.brand}
              onChange={(e) => set('brand', e.target.value)}
              list="brand-options"
              className="mt-1 h-9 w-full rounded border border-gray-300 px-2"
            />
            <datalist id="brand-options">
              {brands?.map((b) => <option key={b} value={b} />)}
            </datalist>
          </label>
          <label className="col-span-1">
            Net Gramaj
            <input
              type="number"
              step="any"
              value={form.net_weight ?? ''}
              onChange={(e) => set('net_weight', e.target.value === '' ? null : Number(e.target.value))}
              className="mt-1 h-9 w-full rounded border border-gray-300 px-2"
            />
          </label>
          <label className="col-span-1">
            Birim
            <input
              value={form.unit}
              onChange={(e) => set('unit', e.target.value)}
              placeholder="g, kg, ml, l, adet"
              className="mt-1 h-9 w-full rounded border border-gray-300 px-2"
            />
          </label>
          <label className="col-span-1">
            Fiyat (₺)
            <input
              type="number"
              step="any"
              value={form.price}
              onChange={(e) => set('price', Number(e.target.value))}
              className="mt-1 h-9 w-full rounded border border-gray-300 px-2"
            />
          </label>
          <label className="col-span-1">
            KDV %
            <input
              type="number"
              step="any"
              value={form.vat_rate ?? ''}
              onChange={(e) => set('vat_rate', e.target.value === '' ? null : Number(e.target.value))}
              className="mt-1 h-9 w-full rounded border border-gray-300 px-2"
            />
          </label>
          <label className="col-span-2">
            Açıklama
            <textarea
              value={form.description}
              onChange={(e) => set('description', e.target.value)}
              rows={2}
              className="mt-1 w-full rounded border border-gray-300 px-2 py-1"
            />
          </label>
          <label className="col-span-2 flex items-center gap-2">
            <input
              type="checkbox"
              checked={form.is_active}
              onChange={(e) => set('is_active', e.target.checked)}
            />
            Aktif
          </label>
        </div>

        <div className="mt-6 flex gap-2">
          <button
            onClick={onClose}
            className="h-10 flex-1 rounded border border-gray-300 text-sm font-medium"
          >
            Vazgeç
          </button>
          <button
            onClick={() => save.mutate()}
            disabled={save.isPending || !form.sku || !form.name || !form.category}
            className="flex h-10 flex-1 items-center justify-center rounded bg-gray-900 text-sm font-medium text-white disabled:opacity-50"
          >
            {save.isPending ? <Spinner /> : 'Kaydet'}
          </button>
        </div>
      </div>
    </div>
  )
}

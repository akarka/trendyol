import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { deleteProduct, getProducts, Product } from '../api/products'
import { Spinner } from '../components/Spinner'
import { ProductFormModal } from '../components/ProductFormModal'
import { useToast } from '../context/ToastContext'

export function ProductsPage() {
  const [search, setSearch] = useState('')
  const [editing, setEditing] = useState<Product | null>(null)
  const [creating, setCreating] = useState(false)
  const toast = useToast()
  const qc = useQueryClient()

  const { data, isLoading, isError } = useQuery({ queryKey: ['products'], queryFn: () => getProducts(false) })

  const remove = useMutation({
    mutationFn: (sku: string) => deleteProduct(sku),
    onSuccess: () => {
      toast.show('success', 'Ürün silindi')
      qc.invalidateQueries({ queryKey: ['products'] })
    },
    onError: (e: any) => toast.show('error', e?.response?.data?.error ?? 'Silinemedi'),
  })

  const products = (data ?? []).filter((p) => {
    const q = search.trim().toLowerCase()
    if (!q) return true
    return (
      p.sku.toLowerCase().includes(q) ||
      p.name.toLowerCase().includes(q) ||
      p.barcode.toLowerCase().includes(q) ||
      p.category.toLowerCase().includes(q)
    )
  })

  const onDelete = (p: Product) => {
    if (!window.confirm(`"${p.name}" (${p.sku}) silinsin mi?`)) return
    remove.mutate(p.sku)
  }

  return (
    <div>
      <div className="mb-4 flex items-center justify-between gap-2">
        <h1 className="text-xl font-semibold">Ürünler</h1>
        <button
          onClick={() => setCreating(true)}
          className="h-9 rounded bg-gray-900 px-3 text-sm font-medium text-white"
        >
          + Yeni Ürün
        </button>
      </div>

      <input
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="SKU, ad, barkod veya kategoriye göre ara…"
        className="mb-4 h-10 w-full rounded border border-gray-300 px-3 text-sm sm:max-w-sm"
      />

      {isLoading && (
        <div className="flex justify-center py-10 text-gray-400">
          <Spinner />
        </div>
      )}
      {isError && <div className="py-10 text-center text-red-600">Bağlantı hatası</div>}
      {!isLoading && !isError && products.length === 0 && (
        <div className="py-10 text-center text-gray-400">Ürün yok</div>
      )}

      {products.length > 0 && (
        <>
          {/* Desktop tablo */}
          <table className="hidden w-full overflow-hidden rounded-lg bg-white text-sm shadow sm:table">
            <thead className="bg-gray-50 text-left text-gray-500">
              <tr>
                <th className="px-4 py-2">SKU</th>
                <th className="px-4 py-2">Ürün</th>
                <th className="px-4 py-2">Kategori</th>
                <th className="px-4 py-2">Fiyat</th>
                <th className="px-4 py-2">Durum</th>
                <th className="px-4 py-2"></th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {products.map((p) => (
                <tr key={p.sku} className="hover:bg-gray-50">
                  <td className="px-4 py-2 font-mono text-xs">{p.sku}</td>
                  <td className="px-4 py-2">
                    {p.name}
                    {p.needs_fix && <span className="ml-2 text-xs text-amber-600">⚠ düzeltilmeli</span>}
                  </td>
                  <td className="px-4 py-2 text-gray-500">{p.category}</td>
                  <td className="px-4 py-2 tabular-nums">{p.price.toFixed(2)} ₺</td>
                  <td className="px-4 py-2">
                    <span className={p.is_active ? 'text-green-600' : 'text-gray-400'}>
                      {p.is_active ? 'Aktif' : 'Pasif'}
                    </span>
                  </td>
                  <td className="px-4 py-2 text-right">
                    <button onClick={() => setEditing(p)} className="mr-3 text-xs text-gray-600 underline">
                      Düzenle
                    </button>
                    <button
                      onClick={() => onDelete(p)}
                      disabled={remove.isPending && remove.variables === p.sku}
                      className="text-xs text-red-600 underline disabled:opacity-50"
                    >
                      Sil
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {/* Mobil kartlar */}
          <div className="grid gap-2 sm:hidden">
            {products.map((p) => (
              <div key={p.sku} className="rounded-lg bg-white p-3 shadow">
                <div className="flex items-center justify-between">
                  <span className="font-medium">{p.name}</span>
                  <span className={p.is_active ? 'text-xs text-green-600' : 'text-xs text-gray-400'}>
                    {p.is_active ? 'Aktif' : 'Pasif'}
                  </span>
                </div>
                <div className="mt-1 text-xs text-gray-500">
                  {p.sku} · {p.category} · {p.price.toFixed(2)} ₺
                </div>
                <div className="mt-2 flex gap-3">
                  <button onClick={() => setEditing(p)} className="text-xs text-gray-600 underline">
                    Düzenle
                  </button>
                  <button onClick={() => onDelete(p)} className="text-xs text-red-600 underline">
                    Sil
                  </button>
                </div>
              </div>
            ))}
          </div>
        </>
      )}

      {creating && <ProductFormModal onClose={() => setCreating(false)} />}
      {editing && <ProductFormModal product={editing} onClose={() => setEditing(null)} />}
    </div>
  )
}

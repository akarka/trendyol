import { useRef, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { downloadBackup, importProducts, restoreBackup } from '../api/admin'
import { useToast } from '../context/ToastContext'

export function BackupPage() {
  const [file, setFile] = useState<File | null>(null)
  const fileInput = useRef<HTMLInputElement>(null)
  const [productFile, setProductFile] = useState<File | null>(null)
  const productFileInput = useRef<HTMLInputElement>(null)
  const [importLog, setImportLog] = useState<string | null>(null)
  const toast = useToast()

  const backup = useMutation({
    mutationFn: downloadBackup,
    onSuccess: () => toast.show('success', 'Yedek indirildi'),
    onError: () => toast.show('error', 'Yedekleme başarısız'),
  })

  const restore = useMutation({
    mutationFn: (f: File) => restoreBackup(f),
    onSuccess: () => {
      toast.show('success', 'Veritabanı geri yüklendi')
      setFile(null)
      if (fileInput.current) fileInput.current.value = ''
    },
    onError: (e: any) => toast.show('error', e?.response?.data?.error ?? 'Geri yükleme başarısız'),
  })

  const importMutation = useMutation({
    mutationFn: (f: File) => importProducts(f),
    onSuccess: (res) => {
      toast.show('success', 'İçe aktarma tamamlandı')
      setImportLog(res.log)
      setProductFile(null)
      if (productFileInput.current) productFileInput.current.value = ''
    },
    onError: (e: any) => {
      toast.show('error', 'İçe aktarma başarısız')
      setImportLog(e?.response?.data?.log ?? e?.response?.data?.error ?? null)
    },
  })

  const onRestore = () => {
    if (!file) return
    if (
      !window.confirm(
        `"${file.name}" dosyası ile veritabanı GERİ YÜKLENECEK. Mevcut tüm veriler (siparişler, kullanıcılar, ayarlar) bu dosyadakiyle DEĞİŞTİRİLECEK ve geri alınamaz. Onaylıyor musunuz?`,
      )
    )
      return
    restore.mutate(file)
  }

  return (
    <div className="space-y-6">
      <h1 className="text-xl font-semibold">Yedekleme</h1>

      <div className="rounded-lg bg-white p-4 shadow">
        <h2 className="mb-1 text-sm font-semibold text-gray-800">Veritabanı Yedeği</h2>
        <p className="mb-3 text-sm text-gray-500">
          Tüm tabloların (siparişler, ürünler, kullanıcılar, ayarlar) tam yedeğini .sql.gz olarak indirir.
        </p>
        <button
          onClick={() => backup.mutate()}
          disabled={backup.isPending}
          className="h-10 rounded bg-gray-900 px-4 text-sm font-medium text-white disabled:opacity-50"
        >
          {backup.isPending ? 'İndiriliyor…' : 'Yedek İndir'}
        </button>
      </div>

      <div className="rounded-lg bg-white p-4 shadow">
        <h2 className="mb-1 text-sm font-semibold text-gray-800">Geri Yükleme</h2>
        <p className="mb-3 text-sm text-red-600">
          Dikkat: seçilen yedek dosyası mevcut tüm verinin üzerine yazar. Bu işlem geri alınamaz.
        </p>
        <div className="flex flex-wrap items-center gap-3">
          <input
            ref={fileInput}
            type="file"
            accept=".sql,.gz"
            onChange={(e) => setFile(e.target.files?.[0] ?? null)}
            className="text-sm"
          />
          <button
            onClick={onRestore}
            disabled={!file || restore.isPending}
            className="h-10 rounded bg-red-600 px-4 text-sm font-medium text-white disabled:opacity-40"
          >
            {restore.isPending ? 'Geri Yükleniyor…' : 'Geri Yükle'}
          </button>
        </div>
      </div>

      <div className="rounded-lg bg-white p-4 shadow">
        <h2 className="mb-1 text-sm font-semibold text-gray-800">Ürün İçe Aktarma (CSV/XLSX)</h2>
        <p className="mb-3 text-sm text-gray-500">
          Katalog şablonundaki (export-products formatı) dosyayı yükle. SKU'su veritabanında zaten
          olan satırlar güncellenmez, atlanır — sadece yeni ürünler eklenir.
        </p>
        <div className="flex flex-wrap items-center gap-3">
          <input
            ref={productFileInput}
            type="file"
            accept=".csv,.xlsx"
            onChange={(e) => setProductFile(e.target.files?.[0] ?? null)}
            className="text-sm"
          />
          <button
            onClick={() => productFile && importMutation.mutate(productFile)}
            disabled={!productFile || importMutation.isPending}
            className="h-10 rounded bg-gray-900 px-4 text-sm font-medium text-white disabled:opacity-40"
          >
            {importMutation.isPending ? 'Aktarılıyor…' : 'İçe Aktar'}
          </button>
        </div>
        {importLog && (
          <pre className="mt-3 max-h-64 overflow-auto rounded bg-gray-50 p-3 text-xs text-gray-600">
            {importLog}
          </pre>
        )}
      </div>
    </div>
  )
}

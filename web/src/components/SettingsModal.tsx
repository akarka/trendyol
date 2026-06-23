import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { getSettings, updateSetting } from '../api/settings'
import { Spinner } from './Spinner'
import { useToast } from '../context/ToastContext'
import { LABEL_LAYOUT_KEY } from '../lib/labelLayout'

function SettingRow({ settingKey, initial }: { settingKey: string; initial: string }) {
  const [value, setValue] = useState(initial)
  const toast = useToast()
  const qc = useQueryClient()

  useEffect(() => setValue(initial), [initial])

  const save = useMutation({
    mutationFn: () => updateSetting(settingKey, value),
    onSuccess: () => {
      toast.show('success', 'Kaydedildi')
      qc.invalidateQueries({ queryKey: ['settings'] })
    },
    onError: () => toast.show('error', 'Hata'),
  })

  return (
    <div className="flex flex-col gap-2 rounded-lg bg-white p-3 shadow sm:flex-row sm:items-center">
      <label className="w-48 shrink-0 text-sm font-medium text-gray-700">{settingKey}</label>
      <input
        value={value}
        onChange={(e) => setValue(e.target.value)}
        className="h-10 w-full rounded border-gray-300 text-sm"
      />
      <button
        onClick={() => save.mutate()}
        disabled={save.isPending || value === initial}
        className="flex h-10 w-24 shrink-0 items-center justify-center rounded bg-gray-900 text-sm font-medium text-white disabled:opacity-40"
      >
        {save.isPending ? <Spinner /> : 'Kaydet'}
      </button>
    </div>
  )
}

// Etiket yerleşimi (kağıt/sütun-satır/hücre boşluğu) artık Manuel Sipariş sayfasında
// "Kağıt Ayarları" ve "Etiket Ayarları" bölümlerinde inline düzenlenir; bu modal sadece
// diğer (etiket yerleşimi dışı) ayarları listeler.
export function SettingsModal({ onClose }: { onClose: () => void }) {
  const { data, isLoading, isError } = useQuery({ queryKey: ['settings'], queryFn: getSettings })

  const keys = data
    ? Object.keys(data)
        .filter((k) => k !== LABEL_LAYOUT_KEY)
        .sort()
    : []

  return (
    <div
      className="fixed inset-0 z-40 flex items-end justify-center bg-black/40 sm:items-center"
      onClick={onClose}
    >
      <div
        className="max-h-[90vh] w-full overflow-y-auto bg-gray-50 p-5 sm:max-w-2xl sm:rounded-lg"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold">Ayarlar</h2>
          <button onClick={onClose} className="text-sm text-gray-500 underline">
            Kapat
          </button>
        </div>

        {isLoading && (
          <div className="flex justify-center py-10 text-gray-400">
            <Spinner />
          </div>
        )}
        {isError && <div className="py-10 text-center text-red-600">Bağlantı hatası</div>}

        {data && (
          <div className="grid gap-2">
            {keys.length === 0 && <div className="py-6 text-center text-gray-400">Ayar yok</div>}
            {keys.map((k) => (
              <SettingRow key={k} settingKey={k} initial={data[k]} />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

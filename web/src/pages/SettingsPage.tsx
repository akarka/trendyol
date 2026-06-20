import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { getSettings, updateSetting } from '../api/settings'
import { Spinner } from '../components/Spinner'
import { useToast } from '../context/ToastContext'

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

export function SettingsPage() {
  const { data, isLoading, isError } = useQuery({ queryKey: ['settings'], queryFn: getSettings })

  if (isLoading)
    return (
      <div className="flex justify-center py-10 text-gray-400">
        <Spinner />
      </div>
    )
  if (isError || !data) return <div className="py-10 text-center text-red-600">Bağlantı hatası</div>

  const keys = Object.keys(data).sort()

  return (
    <div>
      <h1 className="mb-4 text-xl font-semibold">Ayarlar</h1>
      {keys.length === 0 && <div className="py-10 text-center text-gray-400">Ayar yok</div>}
      <div className="grid gap-2">
        {keys.map((k) => (
          <SettingRow key={k} settingKey={k} initial={data[k]} />
        ))}
      </div>
    </div>
  )
}

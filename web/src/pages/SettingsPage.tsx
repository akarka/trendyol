import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { getSettings, updateSetting } from '../api/settings'
import { Spinner } from '../components/Spinner'
import { SheetPreview } from '../components/SheetPreview'
import { useToast } from '../context/ToastContext'
import { DEFAULT_LAYOUT, LABEL_LAYOUT_KEY, LabelLayout, parseLayout } from '../lib/labelLayout'

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

const FIELDS: { key: keyof LabelLayout; label: string }[] = [
  { key: 'pageWidthMm', label: 'Sayfa En (mm)' },
  { key: 'pageHeightMm', label: 'Sayfa Boy (mm)' },
  { key: 'columns', label: 'Sütun' },
  { key: 'rows', label: 'Satır' },
  { key: 'labelWidthMm', label: 'Etiket En (mm)' },
  { key: 'labelHeightMm', label: 'Etiket Boy (mm)' },
  { key: 'marginTopMm', label: 'Üst Kenar (mm)' },
  { key: 'marginLeftMm', label: 'Sol Kenar (mm)' },
  { key: 'gapXMm', label: 'Yatay Boşluk (mm)' },
  { key: 'gapYMm', label: 'Dikey Boşluk (mm)' },
  { key: 'paddingMm', label: 'Hücre İç Boşluk (mm)' },
]

function LabelLayoutEditor({ initial }: { initial: LabelLayout }) {
  const [layout, setLayout] = useState<LabelLayout>(initial)
  const toast = useToast()
  const qc = useQueryClient()

  useEffect(() => setLayout(initial), [initial])

  const save = useMutation({
    mutationFn: () => updateSetting(LABEL_LAYOUT_KEY, JSON.stringify(layout)),
    onSuccess: () => {
      toast.show('success', 'Yerleşim kaydedildi')
      qc.invalidateQueries({ queryKey: ['settings'] })
    },
    onError: () => toast.show('error', 'Hata'),
  })

  const set = (k: keyof LabelLayout, v: string) =>
    setLayout((l) => ({ ...l, [k]: Number(v) || 0 }))

  return (
    <div className="rounded-lg bg-white p-4 shadow">
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-sm font-semibold text-gray-800">Etiket Yerleşimi (A4 sticker)</h2>
        <button
          onClick={() => setLayout(DEFAULT_LAYOUT)}
          className="text-xs text-gray-500 underline"
        >
          Varsayılana dön
        </button>
      </div>

      <div className="flex flex-col gap-6 sm:flex-row">
        <div className="grid flex-1 grid-cols-2 gap-3">
          {FIELDS.map((f) => (
            <label key={f.key} className="text-xs text-gray-600">
              {f.label}
              <input
                type="number"
                step="any"
                value={layout[f.key]}
                onChange={(e) => set(f.key, e.target.value)}
                className="mt-1 h-9 w-full rounded border border-gray-300 px-2 text-sm"
              />
            </label>
          ))}
        </div>

        <div className="flex flex-col items-center gap-2">
          <SheetPreview layout={layout} />
          <span className="text-xs text-gray-400">{layout.columns}×{layout.rows} önizleme</span>
        </div>
      </div>

      <button
        onClick={() => save.mutate()}
        disabled={save.isPending}
        className="mt-4 h-10 rounded bg-gray-900 px-4 text-sm font-medium text-white disabled:opacity-50"
      >
        {save.isPending ? 'Kaydediliyor…' : 'Yerleşimi Kaydet'}
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

  const keys = Object.keys(data)
    .filter((k) => k !== LABEL_LAYOUT_KEY)
    .sort()

  return (
    <div className="space-y-6">
      <h1 className="text-xl font-semibold">Ayarlar</h1>

      <LabelLayoutEditor initial={parseLayout(data)} />

      <div>
        <h2 className="mb-2 text-sm font-semibold text-gray-800">Diğer Ayarlar</h2>
        {keys.length === 0 && <div className="py-6 text-center text-gray-400">Ayar yok</div>}
        <div className="grid gap-2">
          {keys.map((k) => (
            <SettingRow key={k} settingKey={k} initial={data[k]} />
          ))}
        </div>
      </div>
    </div>
  )
}

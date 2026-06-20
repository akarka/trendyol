import { useQuery } from '@tanstack/react-query'
import { getStatus } from '../api/printer'
import { Spinner } from '../components/Spinner'
import { formatDate } from '../lib/format'

const JOB_STYLE: Record<string, string> = {
  success: 'text-green-700',
  failed: 'text-red-700',
  queued: 'text-gray-500',
}

export function PrinterPage() {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['printer-status'],
    queryFn: getStatus,
    refetchInterval: 10_000,
  })

  if (isLoading)
    return (
      <div className="flex justify-center py-10 text-gray-400">
        <Spinner />
      </div>
    )
  if (isError || !data) return <div className="py-10 text-center text-red-600">Bağlantı hatası</div>

  return (
    <div>
      <h1 className="mb-4 text-xl font-semibold">Yazıcı</h1>

      <div className="mb-4 grid grid-cols-2 gap-3 sm:max-w-md">
        <div className="rounded-lg bg-white p-3 shadow">
          <div className="text-xs text-gray-500">Mod</div>
          <div className="font-medium">{data.test_mode ? 'TEST (txt)' : 'ESC/POS'}</div>
        </div>
        <div className="rounded-lg bg-white p-3 shadow">
          <div className="text-xs text-gray-500">Cihaz</div>
          <div className="truncate font-medium">{data.device}</div>
        </div>
      </div>

      <div className="mb-2 text-sm font-medium text-gray-700">Son baskılar</div>
      <div className="overflow-hidden rounded-lg bg-white shadow">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 text-left text-gray-500">
            <tr>
              <th className="px-4 py-2">Sipariş</th>
              <th className="px-4 py-2">Durum</th>
              <th className="px-4 py-2">Hata</th>
              <th className="px-4 py-2">Zaman</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {data.jobs.length === 0 && (
              <tr>
                <td colSpan={4} className="px-4 py-6 text-center text-gray-400">
                  Kayıt yok
                </td>
              </tr>
            )}
            {data.jobs.map((j) => (
              <tr key={j.id} className={j.status === 'failed' ? 'bg-red-50' : ''}>
                <td className="px-4 py-2 font-medium">{j.order_id}</td>
                <td className={`px-4 py-2 font-medium ${JOB_STYLE[j.status] ?? ''}`}>{j.status}</td>
                <td className="px-4 py-2 text-gray-500">{j.error_msg ?? '—'}</td>
                <td className="px-4 py-2 text-gray-500">{formatDate(j.attempted_at)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

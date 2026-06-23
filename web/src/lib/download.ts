import { AxiosResponse } from 'axios'

// Axios blob response'undan Content-Disposition dosya adını çıkarıp tarayıcıda indirir.
export function downloadBlobResponse(res: AxiosResponse<Blob>, fallbackName: string) {
  const disposition = res.headers['content-disposition'] as string | undefined
  const match = disposition?.match(/filename="?([^"]+)"?/)
  const filename = match?.[1] ?? fallbackName

  const url = URL.createObjectURL(res.data)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}

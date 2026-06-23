import client from './client'
import { downloadBlobResponse } from '../lib/download'

export async function downloadBackup(): Promise<void> {
  const res = await client.get('/api/admin/backup', { responseType: 'blob' })
  downloadBlobResponse(res, 'trendyol-backup.sql.gz')
}

export async function restoreBackup(file: File): Promise<void> {
  const form = new FormData()
  form.append('file', file)
  await client.post('/api/admin/restore', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export interface ImportResult {
  status: string
  log: string
}

export async function importProducts(file: File): Promise<ImportResult> {
  const form = new FormData()
  form.append('file', file)
  const { data } = await client.post<ImportResult>('/api/products/import', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return data
}

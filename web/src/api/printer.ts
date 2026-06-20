import client from './client'

export interface PrintJob {
  id: number
  order_id: string
  status: string
  error_msg: string | null
  attempted_at: string
}

export interface PrinterStatus {
  test_mode: boolean
  device: string
  jobs: PrintJob[]
}

export async function getStatus(): Promise<PrinterStatus> {
  const { data } = await client.get<PrinterStatus>('/api/printer/status')
  return data
}

export async function getLogs(): Promise<PrintJob[]> {
  const { data } = await client.get<PrintJob[]>('/api/logs')
  return data
}

import client from './client'
import { downloadBlobResponse } from '../lib/download'

export interface OrderLine {
  productName: string
  barcode: string
  quantity: number
  amount: number
}

export interface ShipmentAddress {
  firstName?: string
  lastName?: string
  address1?: string
  city?: string
  district?: string
  postalCode?: string
}

export interface OrderPayload {
  id?: string
  orderNumber?: string
  packageStatus?: string
  orderDate?: string
  cargoProviderName?: string
  lines?: OrderLine[]
  shipmentAddress?: ShipmentAddress
}

export interface OrderRow {
  uuid: string
  order_id: string
  order_number: string
  package_status: string
  payload: OrderPayload
  created_at: string
  updated_at: string
}

export interface OrderFilters {
  status?: string
  limit?: number
  offset?: number
}

export async function getOrders(filters: OrderFilters = {}): Promise<OrderRow[]> {
  const { status = '', limit = 50, offset = 0 } = filters
  const { data } = await client.get<OrderRow[]>('/api/orders', {
    params: { status: status || undefined, limit, offset },
  })
  return data
}

export async function getOrder(orderID: string): Promise<OrderRow> {
  const { data } = await client.get<OrderRow>(`/api/orders/${orderID}`)
  return data
}

export async function printOrder(orderID: string): Promise<{ job_id: number; status: string }> {
  const { data } = await client.post(`/api/orders/${orderID}/print`)
  return data
}

export async function exportOrders(status: string): Promise<void> {
  const res = await client.get('/api/orders/export', {
    params: { status: status || undefined },
    responseType: 'blob',
  })
  downloadBlobResponse(res, 'siparisler.xlsx')
}

export interface ManualOrderLine {
  sku: string
  quantity: number
}

export interface ManualOrderResult {
  order_id: string
  order_number: string
  job_id: number
}

export async function createManualOrder(
  customerName: string,
  lines: ManualOrderLine[],
): Promise<ManualOrderResult> {
  const { data } = await client.post<ManualOrderResult>('/api/orders/manual', {
    customer_name: customerName,
    lines,
  })
  return data
}

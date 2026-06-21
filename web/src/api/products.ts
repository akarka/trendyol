import client from './client'

export interface Product {
  sku: string
  barcode: string
  name: string
  marketplace_name: string | null
  category: string
  brand: string | null
  net_weight: number | null
  unit: string | null
  price: number
  vat_rate: number | null
  is_active: boolean
  needs_fix: boolean
  description: string | null
}

export async function getProducts(activeOnly = false): Promise<Product[]> {
  const { data } = await client.get<Product[]>('/api/products', {
    params: activeOnly ? { active: 1 } : {},
  })
  return data
}

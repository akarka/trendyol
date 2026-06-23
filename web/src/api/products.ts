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

export async function getCategories(): Promise<string[]> {
  const { data } = await client.get<string[]>('/api/categories')
  return data
}

export async function getBrands(): Promise<string[]> {
  const { data } = await client.get<string[]>('/api/brands')
  return data
}

export interface ProductInput {
  sku: string
  barcode: string
  name: string
  marketplace_name: string
  category: string
  brand: string
  net_weight: number | null
  unit: string
  price: number
  vat_rate: number | null
  is_active: boolean
  description: string
}

export async function createProduct(input: ProductInput): Promise<void> {
  await client.post('/api/products', input)
}

export async function updateProduct(sku: string, input: ProductInput): Promise<void> {
  await client.put(`/api/products/${sku}`, input)
}

export async function deleteProduct(sku: string): Promise<void> {
  await client.delete(`/api/products/${sku}`)
}

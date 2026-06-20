import client from './client'

export type Settings = Record<string, string>

export async function getSettings(): Promise<Settings> {
  const { data } = await client.get<Settings>('/api/settings')
  return data
}

export async function updateSetting(key: string, value: string): Promise<void> {
  await client.put(`/api/settings/${key}`, { value })
}

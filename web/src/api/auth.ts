import client from './client'

export interface LoginResponse {
  token: string
  role: string
}

export async function login(username: string, password: string): Promise<LoginResponse> {
  const { data } = await client.post<LoginResponse>('/api/auth/login', { username, password })
  return data
}

export interface AuthConfig {
  rbac_enabled: boolean
}

export async function getAuthConfig(): Promise<AuthConfig> {
  const { data } = await client.get<AuthConfig>('/api/config')
  return data
}

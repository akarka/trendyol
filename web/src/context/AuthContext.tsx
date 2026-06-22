import { createContext, useContext, useEffect, useState, ReactNode } from 'react'
import { getAuthConfig } from '../api/auth'

interface AuthState {
  token: string | null
  role: string | null
  rbacEnabled: boolean | null // null = config yükleniyor
  login: (token: string, role: string) => void
  logout: () => void
}

const AuthContext = createContext<AuthState | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'))
  const [role, setRole] = useState<string | null>(() => localStorage.getItem('role'))
  const [rbacEnabled, setRbacEnabled] = useState<boolean | null>(null)

  useEffect(() => {
    getAuthConfig()
      .then((c) => setRbacEnabled(c.rbac_enabled))
      .catch(() => setRbacEnabled(true)) // hata → güvenli taraf: auth açık
  }, [])

  const login = (t: string, r: string) => {
    localStorage.setItem('token', t)
    localStorage.setItem('role', r)
    setToken(t)
    setRole(r)
  }

  const logout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('role')
    setToken(null)
    setRole(null)
  }

  return (
    <AuthContext.Provider value={{ token, role, rbacEnabled, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth, AuthProvider içinde kullanılmalı')
  return ctx
}

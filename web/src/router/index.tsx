import { Navigate, Route, Routes } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { Layout } from '../components/Layout'
import { LoginPage } from '../pages/LoginPage'
import { OrdersPage } from '../pages/OrdersPage'
import { ManualOrderPage } from '../pages/ManualOrderPage'
import { PrinterPage } from '../pages/PrinterPage'
import { SettingsPage } from '../pages/SettingsPage'
import { ReactNode } from 'react'

function ProtectedRoute({ children }: { children: ReactNode }) {
  const { token, rbacEnabled } = useAuth()
  if (rbacEnabled === null) return null // config yükleniyor
  if (!rbacEnabled) return <>{children}</> // auth kapalı → doğrudan erişim
  if (!token) return <Navigate to="/login" replace />
  return <>{children}</>
}

export function AppRouter() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route path="/orders" element={<OrdersPage />} />
        <Route path="/manual" element={<ManualOrderPage />} />
        <Route path="/printer" element={<PrinterPage />} />
        <Route path="/settings" element={<SettingsPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/orders" replace />} />
    </Routes>
  )
}

import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

const NAV = [
  { to: '/orders', label: 'Siparişler', icon: '📦' },
  { to: '/manual', label: 'Manuel Sipariş', icon: '✍️' },
  { to: '/products', label: 'Ürünler', icon: '🛒' },
  { to: '/printer', label: 'Yazıcı', icon: '🖨️' },
  { to: '/settings', label: 'Ayarlar', icon: '⚙️' },
  { to: '/backup', label: 'Yedekleme', icon: '🗄️' },
]

export function Layout() {
  const { logout, rbacEnabled } = useAuth()
  const navigate = useNavigate()

  const onLogout = () => {
    logout()
    navigate('/login', { replace: true })
  }

  const showLogout = rbacEnabled !== false

  return (
    <div className="flex min-h-full">
      {/* Desktop sidebar */}
      <aside className="hidden w-56 shrink-0 flex-col bg-white shadow sm:flex">
        <div className="px-4 py-5 text-lg font-semibold">Yazıcı Yönetimi</div>
        <nav className="flex-1 space-y-1 px-2">
          {NAV.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `flex items-center gap-3 rounded px-3 py-2 text-sm font-medium ${
                  isActive ? 'bg-gray-900 text-white' : 'text-gray-700 hover:bg-gray-100'
                }`
              }
            >
              <span>{item.icon}</span>
              {item.label}
            </NavLink>
          ))}
        </nav>
        {showLogout && (
          <button
            onClick={onLogout}
            className="m-2 rounded px-3 py-2 text-left text-sm text-gray-600 hover:bg-gray-100"
          >
            Çıkış
          </button>
        )}
      </aside>

      {/* Content */}
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex items-center justify-between bg-white px-4 py-3 shadow sm:hidden">
          <span className="font-semibold">Yazıcı Yönetimi</span>
          {showLogout && (
            <button onClick={onLogout} className="text-sm text-gray-600">
              Çıkış
            </button>
          )}
        </header>

        <main className="flex-1 overflow-y-auto p-4 pb-20 sm:pb-4">
          <Outlet />
        </main>

        {/* Mobile bottom nav */}
        <nav className="fixed inset-x-0 bottom-0 flex border-t bg-white sm:hidden">
          {NAV.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `flex flex-1 flex-col items-center gap-0.5 py-2 text-xs ${
                  isActive ? 'text-gray-900' : 'text-gray-400'
                }`
              }
            >
              <span className="text-lg">{item.icon}</span>
              {item.label}
            </NavLink>
          ))}
        </nav>
      </div>
    </div>
  )
}

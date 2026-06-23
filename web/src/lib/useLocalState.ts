import { useEffect, useState } from 'react'

// useState gibi davranır, ama değeri localStorage'da tutar — sayfa/sekme değişse veya
// yenilense de (Manuel Sipariş taslağı, Yeni Ürün formu gibi) state kaybolmaz.
export function useLocalState<T>(key: string, initial: T): [T, React.Dispatch<React.SetStateAction<T>>] {
  const [state, setState] = useState<T>(() => {
    try {
      const raw = localStorage.getItem(key)
      return raw !== null ? (JSON.parse(raw) as T) : initial
    } catch {
      return initial
    }
  })

  useEffect(() => {
    try {
      localStorage.setItem(key, JSON.stringify(state))
    } catch {
      // quota/erişim hatası — sessiz geç
    }
  }, [key, state])

  return [state, setState]
}

export function clearLocalState(key: string): void {
  try {
    localStorage.removeItem(key)
  } catch {
    // yoksay
  }
}

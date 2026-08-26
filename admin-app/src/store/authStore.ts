import { create } from 'zustand'
import { getStoredItem, setStoredItem, removeStoredItem } from '@/lib/storage'

interface AuthState {
  token: string | null
  isAuthenticated: boolean
  setToken: (token: string | null) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: getStoredItem('admin_token'),
  isAuthenticated: !!getStoredItem('admin_token'),
  setToken: (token) => {
    if (token) {
      setStoredItem('admin_token', token)
    } else {
      removeStoredItem('admin_token')
    }
    set({ token, isAuthenticated: !!token })
  },
  logout: () => {
    removeStoredItem('admin_token')
    set({ token: null, isAuthenticated: false })
  },
}))

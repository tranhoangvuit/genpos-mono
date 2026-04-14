import { create } from 'zustand'

import type { AuthUser } from '@/gen/genpos/v1/genpos_pb'

interface AuthState {
  user: AuthUser | null
  setUser: (user: AuthUser | null) => void
  clear: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  setUser: (user) => set({ user }),
  clear: () => set({ user: null }),
}))

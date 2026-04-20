import { useMutation } from '@tanstack/react-query'
import { useNavigate } from '@tanstack/react-router'

import { authClient } from '@/shared/api/client'

import { resetAuthBootstrap } from './bootstrap'
import { useAuthStore } from './store'

export function useSignIn() {
  const navigate = useNavigate()
  const setUser = useAuthStore((s) => s.setUser)
  return useMutation({
    mutationFn: async (input: {
      email: string
      password: string
      rememberMe: boolean
    }) => authClient.signIn(input),
    onSuccess: (res) => {
      setUser(res.user ?? null)
      const slug = res.user?.orgSlug
      if (slug) {
        void navigate({
          to: '/$subdomain/dashboard',
          params: { subdomain: slug },
        })
      }
    },
  })
}

export function useSignUp() {
  const navigate = useNavigate()
  const setUser = useAuthStore((s) => s.setUser)
  return useMutation({
    mutationFn: async (input: {
      domain: string
      email: string
      password: string
    }) => authClient.signUp(input),
    onSuccess: (res) => {
      setUser(res.user ?? null)
      const slug = res.user?.orgSlug
      if (slug) {
        void navigate({
          to: '/$subdomain/dashboard',
          params: { subdomain: slug },
        })
      }
    },
  })
}

export function useSignOut() {
  const navigate = useNavigate()
  return useMutation({
    mutationFn: async () => authClient.signOut({}),
    onSettled: () => {
      resetAuthBootstrap()
      void navigate({ to: '/signin' })
    },
  })
}

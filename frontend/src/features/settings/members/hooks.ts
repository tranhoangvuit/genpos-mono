import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { memberClient } from '@/shared/api/client'

const MEMBERS_KEY = ['settings', 'members'] as const
const ROLES_KEY = ['settings', 'members', 'roles'] as const

export function useMembers() {
  return useQuery({
    queryKey: MEMBERS_KEY,
    queryFn: async () => {
      const res = await memberClient.listMembers({})
      return res.members
    },
  })
}

export function useRoleOptions() {
  return useQuery({
    queryKey: ROLES_KEY,
    queryFn: async () => {
      const res = await memberClient.listRoles({})
      return res.roles
    },
  })
}

export function useCreateMember() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof memberClient.createMember>[0]) =>
      memberClient.createMember(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: MEMBERS_KEY })
    },
  })
}

export function useUpdateMember() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof memberClient.updateMember>[0]) =>
      memberClient.updateMember(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: MEMBERS_KEY })
    },
  })
}

export function useDeleteMember() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => memberClient.deleteMember({ id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: MEMBERS_KEY })
    },
  })
}

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { memberClient } from '@/shared/api/client'

const STAFFS_KEY = ['settings', 'staffs'] as const
const ROLES_KEY = ['settings', 'staffs', 'roles'] as const

export function useStaffs() {
  return useQuery({
    queryKey: STAFFS_KEY,
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

export function useCreateStaff() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof memberClient.createMember>[0]) =>
      memberClient.createMember(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: STAFFS_KEY })
    },
  })
}

export function useUpdateStaff() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (req: Parameters<typeof memberClient.updateMember>[0]) =>
      memberClient.updateMember(req),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: STAFFS_KEY })
    },
  })
}

export function useDeleteStaff() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => memberClient.deleteMember({ id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: STAFFS_KEY })
    },
  })
}

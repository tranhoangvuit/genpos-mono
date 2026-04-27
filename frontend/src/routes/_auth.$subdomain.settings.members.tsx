import { Navigate, createFileRoute } from '@tanstack/react-router'

// Legacy redirect: /settings/members → /settings/staffs.
export const Route = createFileRoute('/_auth/$subdomain/settings/members')({
  component: function RedirectToStaffs() {
    return (
      <Navigate
        to="/$subdomain/settings/staffs"
        params={{ subdomain: Route.useParams().subdomain }}
        replace
      />
    )
  },
})

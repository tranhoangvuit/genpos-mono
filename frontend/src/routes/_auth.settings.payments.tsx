import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

export const Route = createFileRoute('/_auth/settings/payments')({
  component: PaymentsPage,
})

function PaymentsPage() {
  const { t } = useTranslation()
  return (
    <div>
      <h1 className="text-2xl font-semibold">{t('nav.payments')}</h1>
    </div>
  )
}

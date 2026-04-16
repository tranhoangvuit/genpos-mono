import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

export const Route = createFileRoute('/_auth/settings/taxes')({
  component: TaxesPage,
})

function TaxesPage() {
  const { t } = useTranslation()
  return (
    <div>
      <h1 className="text-2xl font-semibold">{t('nav.taxes')}</h1>
    </div>
  )
}

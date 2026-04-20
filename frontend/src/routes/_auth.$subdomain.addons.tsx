import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

export const Route = createFileRoute('/_auth/$subdomain/addons')({
  component: AddonsPage,
})

function AddonsPage() {
  const { t } = useTranslation()
  return (
    <div>
      <h1 className="text-2xl font-semibold">{t('nav.addons')}</h1>
    </div>
  )
}

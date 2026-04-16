import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

export const Route = createFileRoute('/_auth/stock-takes')({
  component: StockTakesPage,
})

function StockTakesPage() {
  const { t } = useTranslation()
  return (
    <div>
      <h1 className="text-2xl font-semibold">{t('nav.stockTakes')}</h1>
    </div>
  )
}

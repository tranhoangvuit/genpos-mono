import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

export const Route = createFileRoute('/_auth/daily-sales-report')({
  component: DailySalesReportPage,
})

function DailySalesReportPage() {
  const { t } = useTranslation()
  return (
    <div>
      <h1 className="text-2xl font-semibold">{t('nav.dailySalesReport')}</h1>
    </div>
  )
}

import { createFileRoute } from '@tanstack/react-router'

import { DailySalesReportPage } from '@/features/reports/DailySalesReportPage'

export const Route = createFileRoute('/_auth/daily-sales-report/')({
  component: DailySalesReportPage,
})

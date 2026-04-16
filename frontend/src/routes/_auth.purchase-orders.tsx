import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

export const Route = createFileRoute('/_auth/purchase-orders')({
  component: PurchaseOrdersPage,
})

function PurchaseOrdersPage() {
  const { t } = useTranslation()
  return (
    <div>
      <h1 className="text-2xl font-semibold">{t('nav.purchaseOrders')}</h1>
    </div>
  )
}

import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

export const Route = createFileRoute('/_auth/products/categories')({
  component: CategoriesPage,
})

function CategoriesPage() {
  const { t } = useTranslation()
  return (
    <div>
      <h1 className="text-2xl font-semibold">{t('nav.categories')}</h1>
    </div>
  )
}

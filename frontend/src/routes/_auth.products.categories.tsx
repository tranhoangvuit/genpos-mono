import { createFileRoute } from '@tanstack/react-router'

import { CategoriesPage } from '@/features/catalog/CategoriesPage'

export const Route = createFileRoute('/_auth/products/categories')({
  component: CategoriesPage,
})

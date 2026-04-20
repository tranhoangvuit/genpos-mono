import { createFileRoute } from '@tanstack/react-router'

import { ProductFormPage } from '@/features/catalog/ProductFormPage'

export const Route = createFileRoute('/_auth/$subdomain/products/new')({
  component: () => <ProductFormPage mode="create" />,
})

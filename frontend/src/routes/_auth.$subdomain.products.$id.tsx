import { createFileRoute, useParams } from '@tanstack/react-router'

import { ProductFormPage } from '@/features/catalog/ProductFormPage'

export const Route = createFileRoute('/_auth/$subdomain/products/$id')({
  component: EditRoute,
})

function EditRoute() {
  const { id } = useParams({ from: '/_auth/$subdomain/products/$id' })
  return <ProductFormPage mode="edit" productId={id} />
}

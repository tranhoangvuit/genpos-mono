import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/')({
  component: Home,
})

function Home() {
  return (
    <div style={{ padding: '2rem', fontFamily: 'system-ui, sans-serif' }}>
      <h1>GenPOS</h1>
      <p>Backend: ConnectRPC on Go</p>
      <p>Frontend: TanStack Start</p>
      <p>Infrastructure: PostgreSQL 17 + Redis + PowerSync</p>
    </div>
  )
}

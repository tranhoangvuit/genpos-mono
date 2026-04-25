import { serve } from 'srvx/node'
import { serveStatic } from 'srvx/static'
import handler from './dist/server/server.js'

const port = Number(process.env.PORT) || 3032
const hostname = process.env.HOST || '0.0.0.0'

serve({
  port,
  hostname,
  middleware: [serveStatic({ dir: 'dist/client' })],
  fetch: (req) => handler.fetch(req),
})

console.log(`frontend listening on http://${hostname}:${port}`)

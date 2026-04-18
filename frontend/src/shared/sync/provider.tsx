import { PowerSyncContext } from '@powersync/react'
import { useEffect, useState, type ReactNode } from 'react'

import { connectSync, getSyncDB } from './client'

type Props = { children: ReactNode }

// SyncProvider ensures the PowerSync database is connected and exposes it
// via PowerSyncContext so descendants can use @powersync/react's useQuery.
export function SyncProvider({ children }: Props) {
  const [db, setDb] = useState(getSyncDB())

  useEffect(() => {
    let cancelled = false
    void connectSync().then(() => {
      if (!cancelled) setDb(getSyncDB())
    })
    return () => {
      cancelled = true
    }
  }, [])

  if (!db) {
    return (
      <div className="flex h-svh w-full items-center justify-center text-sm text-[color:var(--color-muted-foreground)]">
        Connecting...
      </div>
    )
  }
  return <PowerSyncContext.Provider value={db}>{children}</PowerSyncContext.Provider>
}

import {
  PowerSyncDatabase,
  type AbstractPowerSyncDatabase,
  type PowerSyncBackendConnector,
} from '@powersync/web'

import { authClient } from '@/shared/api/client'

import { appSchema } from './schema'

let db: AbstractPowerSyncDatabase | null = null

class Connector implements PowerSyncBackendConnector {
  async fetchCredentials() {
    const res = await authClient.getSyncToken({})
    return {
      endpoint: res.endpoint,
      token: res.token,
      expiresAt: new Date(Number(res.expiresAt) * 1000),
    }
  }

  async uploadData(): Promise<void> {
    // Read-only MVP — write path not implemented yet
  }
}

export async function connectSync(): Promise<void> {
  if (db) return

  db = new PowerSyncDatabase({
    schema: appSchema,
    database: { dbFilename: 'genpos.db' },
  })
  await db.init()
  await db.connect(new Connector())
}

export async function disconnectSync(): Promise<void> {
  if (!db) return
  await db.disconnect()
  db = null
}

export function getSyncDB(): AbstractPowerSyncDatabase | null {
  return db
}

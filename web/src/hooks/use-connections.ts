import { useCallback, useEffect, useState } from 'react'
import { api } from '../api/client'
import type { Connection } from '../types'

export function useConnections() {
  const [connections, setConnections] = useState<Connection[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      setConnections(await api.connections.list())
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

  const createConnection = async (data: { name: string; db_type: string; dsn: string }) => {
    const conn = await api.connections.create(data)
    setConnections((prev) => [...prev, conn])
    return conn
  }

  const deleteConnection = async (id: number) => {
    await api.connections.delete(id)
    setConnections((prev) => prev.filter((c) => c.id !== id))
  }

  const testConnection = async (id: number) => {
    return api.connections.test(id)
  }

  return {
    connections,
    loading,
    error,
    createConnection,
    deleteConnection,
    testConnection,
    reload: load,
  }
}

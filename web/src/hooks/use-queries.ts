import { useCallback, useEffect, useState } from 'react'
import { api } from '../api/client'
import type { QueryDelta } from '../types'
import { useSSE } from './use-sse'

export function useQueries(connId: number) {
  const [queries, setQueries] = useState<QueryDelta[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null)

  const load = useCallback(async () => {
    try {
      setLoading(true)
      setError(null)
      const data = await api.queries.list(connId)
      setQueries(data)
      setLastUpdated(new Date())
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setLoading(false)
    }
  }, [connId])

  useEffect(() => {
    load()
  }, [load])

  const sseData = useSSE<QueryDelta[]>(api.queries.streamUrl(connId))
  useEffect(() => {
    if (sseData) {
      setQueries(sseData)
      setLastUpdated(new Date())
    }
  }, [sseData])

  return { queries, loading, error, lastUpdated, reload: load }
}

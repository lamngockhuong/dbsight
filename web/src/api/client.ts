import type { ApiError, Connection, QueryDelta, QuerySnapshot, SlowQuery } from '../types'

const BASE = import.meta.env.VITE_API_BASE ?? ''

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}/api${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  const data = await res.json()
  if (!res.ok) throw new Error((data as ApiError).error ?? 'request failed')
  return data as T
}

export const api = {
  connections: {
    list: () => request<Connection[]>('/connections'),
    create: (body: { name: string; db_type: string; dsn: string }) =>
      request<Connection>('/connections', { method: 'POST', body: JSON.stringify(body) }),
    get: (id: number) => request<Connection>(`/connections/${id}`),
    update: (id: number, body: { name?: string; db_type?: string; dsn?: string }) =>
      request<Connection>(`/connections/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
    delete: (id: number) => request<void>(`/connections/${id}`, { method: 'DELETE' }),
    test: (id: number) =>
      request<{ latency_ms: number }>(`/connections/${id}/test`, { method: 'POST' }),
  },
  queries: {
    list: (connId: number) => request<QueryDelta[]>(`/connections/${connId}/queries`),
    history: (connId: number, limit = 20) =>
      request<QuerySnapshot[]>(`/connections/${connId}/queries/history?limit=${limit}`),
    streamUrl: (connId: number) => `${BASE}/api/connections/${connId}/queries/stream`,
  },
  paste: {
    parseQueries: (logText: string) =>
      request<SlowQuery[]>('/paste/queries', {
        method: 'POST',
        body: JSON.stringify({ log_text: logText }),
      }),
  },
}

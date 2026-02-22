import { useEffect, useState } from 'react'
import { Line, LineChart, ResponsiveContainer, Tooltip } from 'recharts'
import { api } from '@/api/client'

interface SparklinePoint {
  captured_at: string
  mean_exec_ms: number
}

interface QuerySparklineProps {
  queryId: string
  connId: number
}

export function QuerySparkline({ queryId, connId }: QuerySparklineProps) {
  const [points, setPoints] = useState<SparklinePoint[]>([])

  useEffect(() => {
    api.queries
      .history(connId)
      .then((snaps) => {
        const pts = snaps
          .map((s) => ({
            captured_at: s.captured_at,
            mean_exec_ms: s.queries.find((q) => q.query_id === queryId)?.mean_exec_ms ?? 0,
          }))
          .reverse()
        setPoints(pts)
      })
      .catch(() => {})
  }, [queryId, connId])

  if (points.length < 2) return <p className="text-xs text-muted-foreground">Not enough data</p>

  return (
    <ResponsiveContainer width="100%" height={80}>
      <LineChart data={points}>
        <Tooltip
          formatter={(v: number | undefined) => [
            v != null ? `${v.toFixed(2)} ms` : '-',
            'Avg Exec',
          ]}
          labelFormatter={() => ''}
        />
        <Line
          type="monotone"
          dataKey="mean_exec_ms"
          dot={false}
          stroke="hsl(var(--primary))"
          strokeWidth={2}
        />
      </LineChart>
    </ResponsiveContainer>
  )
}

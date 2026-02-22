import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { QueryDetailDrawer } from '@/components/queries/query-detail-drawer'
import { SlowQueryTable } from '@/components/queries/slow-query-table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { useQueries } from '@/hooks/use-queries'
import type { QueryDelta } from '@/types'

export function QueriesPage() {
  const { id } = useParams<{ id: string }>()
  const connId = parseInt(id!, 10)
  const { queries, loading, error, lastUpdated, reload } = useQueries(connId)
  const [selected, setSelected] = useState<QueryDelta | null>(null)
  const navigate = useNavigate()

  if (loading) return <p className="text-muted-foreground">Loading queries...</p>
  if (error) return <p className="text-destructive">{error}</p>

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Slow Queries</h1>
        <div className="flex items-center gap-3">
          {lastUpdated && (
            <Badge variant="outline" className="text-xs">
              Updated: {lastUpdated.toLocaleTimeString()}
            </Badge>
          )}
          <Button size="sm" variant="outline" onClick={reload}>
            Refresh
          </Button>
        </div>
      </div>
      <SlowQueryTable data={queries} onSelect={setSelected} />
      <QueryDetailDrawer
        query={selected}
        connId={connId}
        onClose={() => setSelected(null)}
        onExplain={(q) => navigate(`/explain/${connId}`, { state: { query: q } })}
      />
    </div>
  )
}

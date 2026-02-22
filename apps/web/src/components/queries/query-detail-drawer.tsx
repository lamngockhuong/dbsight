import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { QueryDelta } from '@/types'
import { QuerySparkline } from './query-sparkline'

interface QueryDetailDrawerProps {
  query: QueryDelta | null
  connId: number
  onClose: () => void
  onExplain: (q: string) => void
}

export function QueryDetailDrawer({ query, connId, onClose, onExplain }: QueryDetailDrawerProps) {
  if (!query) return null

  return (
    <div className="fixed inset-y-0 right-0 w-[480px] bg-background border-l shadow-lg z-50 overflow-y-auto">
      <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold">Query Detail</h3>
          <Button variant="ghost" size="sm" onClick={onClose}>
            ✕
          </Button>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm">SQL</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="text-xs bg-muted p-3 rounded-md overflow-x-auto whitespace-pre-wrap">
              {query.query}
            </pre>
          </CardContent>
        </Card>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-sm text-muted-foreground">Calls</p>
            <p className="text-lg font-semibold">
              {query.calls} <Badge variant="secondary">+{query.calls_delta}</Badge>
            </p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Avg Exec</p>
            <p className="text-lg font-semibold">{query.mean_exec_ms.toFixed(2)} ms</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Total Exec</p>
            <p className="text-lg font-semibold">{query.total_exec_ms.toFixed(2)} ms</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Rows</p>
            <p className="text-lg font-semibold">{query.rows}</p>
          </div>
        </div>

        <div>
          <p className="text-sm text-muted-foreground mb-2">Execution Trend</p>
          <QuerySparkline queryId={query.query_id} connId={connId} />
        </div>

        <Button onClick={() => onExplain(query.query)} className="w-full">
          Run EXPLAIN →
        </Button>
      </div>
    </div>
  )
}

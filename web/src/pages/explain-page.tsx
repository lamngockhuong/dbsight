import { useLocation, useParams } from 'react-router-dom'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export function ExplainPage() {
  const { id } = useParams<{ id: string }>()
  const location = useLocation()
  const query = (location.state as { query?: string })?.query

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">EXPLAIN Plan</h1>
      <Card>
        <CardHeader>
          <CardTitle>Connection #{id}</CardTitle>
        </CardHeader>
        <CardContent>
          {query ? (
            <pre className="text-xs bg-muted p-3 rounded-md">{query}</pre>
          ) : (
            <p className="text-muted-foreground">
              Select a query from the Slow Queries page to analyze.
            </p>
          )}
          <p className="text-sm text-muted-foreground mt-4">
            EXPLAIN visualization coming in Phase 08 (post-MVP).
          </p>
        </CardContent>
      </Card>
    </div>
  )
}

import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import type { Recommendation } from '@/types'

const severityVariant: Record<string, 'destructive' | 'secondary' | 'outline'> = {
  high: 'destructive',
  medium: 'secondary',
  low: 'outline',
}

export function RecommendationsList({ items }: { items: Recommendation[] }) {
  if (!items.length) {
    return (
      <p className="text-sm text-green-600 dark:text-green-400">
        No recommendations — looking good!
      </p>
    )
  }

  return (
    <div className="space-y-3">
      {items.map((r, i) => (
        <Card key={i}>
          <CardContent className="pt-4 space-y-2">
            <div className="flex items-center gap-2">
              <Badge variant={severityVariant[r.severity] ?? 'outline'}>
                {r.severity.toUpperCase()}
              </Badge>
              <span className="text-sm">{r.description}</span>
            </div>
            {r.sql && (
              <pre className="text-xs bg-muted p-2 rounded-md overflow-x-auto font-mono">
                {r.sql}
              </pre>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  )
}

import { useParams } from 'react-router-dom'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export function IndexesPage() {
  const { id } = useParams<{ id: string }>()

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">Index Analysis</h1>
      <Card>
        <CardHeader>
          <CardTitle>Connection #{id}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground">Index analysis coming in Phase 09 (post-MVP).</p>
        </CardContent>
      </Card>
    </div>
  )
}

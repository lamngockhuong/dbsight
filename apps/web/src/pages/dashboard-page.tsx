import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useConnections } from '@/hooks/use-connections'

export function DashboardPage() {
  const { connections, loading } = useConnections()
  const navigate = useNavigate()

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Dashboard</h1>
      {loading ? (
        <p className="text-muted-foreground">Loading...</p>
      ) : connections.length === 0 ? (
        <Card>
          <CardContent className="pt-6">
            <p className="text-muted-foreground mb-4">No connections configured yet.</p>
            <Button onClick={() => navigate('/connections')}>Add Connection</Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {connections.map((conn) => (
            <Card
              key={conn.id}
              className="cursor-pointer hover:border-primary transition-colors"
              onClick={() => navigate(`/queries/${conn.id}`)}
            >
              <CardHeader>
                <CardTitle className="text-base">{conn.name}</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">{conn.db_type}</p>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}

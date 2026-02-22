import { ConnectionForm } from '@/components/connections/connection-form'
import { ConnectionList } from '@/components/connections/connection-list'
import { useConnections } from '@/hooks/use-connections'

export function ConnectionsPage() {
  const { connections, loading, error, createConnection, deleteConnection, testConnection } =
    useConnections()

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Connections</h1>
      {error && <p className="text-destructive">{error}</p>}
      <ConnectionForm
        onSubmit={async (data) => {
          await createConnection(data)
        }}
      />
      {loading ? (
        <p className="text-muted-foreground">Loading...</p>
      ) : (
        <ConnectionList
          connections={connections}
          onDelete={deleteConnection}
          onTest={testConnection}
        />
      )}
    </div>
  )
}

import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { Connection } from '@/types'

interface ConnectionListProps {
  connections: Connection[]
  onDelete: (id: number) => Promise<void>
  onTest: (id: number) => Promise<{ latency_ms: number }>
}

export function ConnectionList({ connections, onDelete, onTest }: ConnectionListProps) {
  const navigate = useNavigate()
  const [testResults, setTestResults] = useState<Record<number, string>>({})
  const [testing, setTesting] = useState<Record<number, boolean>>({})

  const handleTest = async (id: number) => {
    setTesting((prev) => ({ ...prev, [id]: true }))
    try {
      const result = await onTest(id)
      setTestResults((prev) => ({ ...prev, [id]: `${result.latency_ms}ms` }))
    } catch (e) {
      setTestResults((prev) => ({ ...prev, [id]: (e as Error).message }))
    } finally {
      setTesting((prev) => ({ ...prev, [id]: false }))
    }
  }

  if (connections.length === 0) {
    return <p className="text-muted-foreground text-sm">No connections yet. Add one above.</p>
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Type</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {connections.map((conn) => (
          <TableRow key={conn.id}>
            <TableCell className="font-medium">{conn.name}</TableCell>
            <TableCell>
              <Badge variant="secondary">{conn.db_type}</Badge>
            </TableCell>
            <TableCell>
              {testResults[conn.id] && (
                <span className="text-sm">
                  {testResults[conn.id].includes('ms') ? (
                    <Badge variant="default">{testResults[conn.id]}</Badge>
                  ) : (
                    <Badge variant="destructive">{testResults[conn.id]}</Badge>
                  )}
                </span>
              )}
            </TableCell>
            <TableCell className="flex gap-2">
              <Button size="sm" variant="outline" onClick={() => navigate(`/queries/${conn.id}`)}>
                Queries
              </Button>
              <Button
                size="sm"
                variant="outline"
                onClick={() => handleTest(conn.id)}
                disabled={testing[conn.id]}
              >
                {testing[conn.id] ? 'Testing...' : 'Test'}
              </Button>
              <Button size="sm" variant="destructive" onClick={() => onDelete(conn.id)}>
                Delete
              </Button>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

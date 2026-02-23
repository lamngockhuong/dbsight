import { useState } from 'react'
import { toast } from 'sonner'
import { api } from '@/api/client'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import type { SlowQuery } from '@/types'

export function PastePage() {
  const [logText, setLogText] = useState('')
  const [queries, setQueries] = useState<SlowQuery[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleParse = async () => {
    setLoading(true)
    setError(null)
    try {
      const result = await api.paste.parseQueries(logText)
      setQueries(result)
      toast.success(`Parsed ${result.length} queries`)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Paste Mode</h1>
      <p className="text-sm text-muted-foreground">
        Paste PostgreSQL slow query log output to analyze offline.
      </p>
      <Card>
        <CardHeader>
          <CardTitle>Slow Query Log</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <Textarea
            value={logText}
            onChange={(e) => setLogText(e.target.value)}
            placeholder={
              'Paste slow query log here...\nduration: 1234.567 ms  statement: SELECT ...'
            }
            rows={10}
          />
          <Button onClick={handleParse} disabled={loading || !logText.trim()}>
            {loading ? 'Parsing...' : 'Analyze'}
          </Button>
          {error && <p className="text-sm text-destructive">{error}</p>}
        </CardContent>
      </Card>
      {queries.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Results ({queries.length} queries)</CardTitle>
          </CardHeader>
          <CardContent className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>#</TableHead>
                  <TableHead>Query</TableHead>
                  <TableHead>Calls</TableHead>
                  <TableHead>Total (ms)</TableHead>
                  <TableHead>Avg (ms)</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {queries.map((q, i) => (
                  <TableRow key={q.query_id}>
                    <TableCell>{i + 1}</TableCell>
                    <TableCell className="max-w-md truncate font-mono text-xs">{q.query}</TableCell>
                    <TableCell>{q.calls}</TableCell>
                    <TableCell>{q.total_exec_ms.toFixed(2)}</TableCell>
                    <TableCell>{q.mean_exec_ms.toFixed(2)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

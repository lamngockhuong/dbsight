import { useCallback, useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '@/api/client'
import { RecommendationsList } from '@/components/indexes/recommendations-list'
import { Badge } from '@/components/ui/badge'
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
import type { IndexAnalysisResult } from '@/types'

function SummaryCard({
  title,
  count,
  severity,
}: {
  title: string
  count: number
  severity: 'high' | 'medium' | 'ok'
}) {
  const colors: Record<string, string> = {
    high: 'border-destructive text-destructive',
    medium: 'border-yellow-500 text-yellow-600 dark:text-yellow-400',
    ok: 'border-green-500 text-green-600 dark:text-green-400',
  }
  return (
    <Card className={`${colors[severity]} border-2`}>
      <CardContent className="pt-4 text-center">
        <div className="text-3xl font-bold">{count}</div>
        <div className="text-sm text-muted-foreground">{title}</div>
      </CardContent>
    </Card>
  )
}

function formatBytes(b: number): string {
  if (b >= 1 << 30) return `${(b / (1 << 30)).toFixed(1)} GB`
  if (b >= 1 << 20) return `${(b / (1 << 20)).toFixed(1)} MB`
  if (b >= 1 << 10) return `${(b / (1 << 10)).toFixed(1)} KB`
  return `${b} B`
}

export function IndexesPage() {
  const { id } = useParams<{ id: string }>()
  const connId = parseInt(id ?? '0', 10)
  const [result, setResult] = useState<IndexAnalysisResult | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    setLoading(true)
    setError(null)
    api.indexes
      .analyze(connId)
      .then(setResult)
      .catch((e) => setError((e as Error).message))
      .finally(() => setLoading(false))
  }, [connId])

  useEffect(() => {
    load()
  }, [load])

  if (loading) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Index Analysis</h1>
        <p className="text-muted-foreground">Analyzing indexes...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Index Analysis</h1>
        <div className="rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </div>
      </div>
    )
  }

  if (!result) return null

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Index Analysis</h1>
        <Button variant="outline" onClick={load}>
          Refresh
        </Button>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
        <SummaryCard
          title="Unused Indexes"
          count={result.unused_indexes.length}
          severity={result.unused_indexes.length > 0 ? 'medium' : 'ok'}
        />
        <SummaryCard
          title="Missing Index Candidates"
          count={result.missing_candidates.length}
          severity={
            result.missing_candidates.length > 5
              ? 'high'
              : result.missing_candidates.length > 0
                ? 'medium'
                : 'ok'
          }
        />
        <SummaryCard
          title="Duplicate Indexes"
          count={result.duplicate_indexes.length}
          severity={result.duplicate_indexes.length > 0 ? 'medium' : 'ok'}
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Recommendations</CardTitle>
        </CardHeader>
        <CardContent>
          <RecommendationsList items={result.recommendations} />
        </CardContent>
      </Card>

      {result.unused_indexes.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              Unused Indexes
              <Badge variant="secondary">{result.unused_indexes.length}</Badge>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Schema</TableHead>
                  <TableHead>Table</TableHead>
                  <TableHead>Index</TableHead>
                  <TableHead className="text-right">Size</TableHead>
                  <TableHead className="text-right">Scans</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {result.unused_indexes.map((idx) => (
                  <TableRow key={`${idx.schema_name}.${idx.index_name}`}>
                    <TableCell>{idx.schema_name}</TableCell>
                    <TableCell>{idx.table_name}</TableCell>
                    <TableCell className="font-mono text-xs">{idx.index_name}</TableCell>
                    <TableCell className="text-right">
                      {formatBytes(idx.index_size_bytes)}
                    </TableCell>
                    <TableCell className="text-right">{idx.index_scans}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {result.missing_candidates.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              Missing Index Candidates
              <Badge variant="secondary">{result.missing_candidates.length}</Badge>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Schema</TableHead>
                  <TableHead>Table</TableHead>
                  <TableHead className="text-right">Seq Scans</TableHead>
                  <TableHead className="text-right">Live Rows</TableHead>
                  <TableHead className="text-right">Table Size</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {result.missing_candidates.map((t) => (
                  <TableRow key={`${t.schema_name}.${t.table_name}`}>
                    <TableCell>{t.schema_name}</TableCell>
                    <TableCell>{t.table_name}</TableCell>
                    <TableCell className="text-right">{t.seq_scans.toLocaleString()}</TableCell>
                    <TableCell className="text-right">{t.n_live_tup.toLocaleString()}</TableCell>
                    <TableCell className="text-right">{formatBytes(t.table_size_bytes)}</TableCell>
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

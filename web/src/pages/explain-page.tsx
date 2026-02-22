import { useState } from 'react'
import { useLocation, useParams } from 'react-router-dom'
import { api } from '@/api/client'
import { ExplainJsonTree } from '@/components/explain/explain-json-tree'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'

export function ExplainPage() {
  const { id } = useParams<{ id: string }>()
  const location = useLocation()
  const connId = parseInt(id ?? '0', 10)
  const initialQuery = (location.state as { query?: string })?.query ?? ''

  const [query, setQuery] = useState(initialQuery)
  const [pasteJSON, setPasteJSON] = useState('')
  const [analyzeMode, setAnalyzeMode] = useState(false)
  const [plan, setPlan] = useState<unknown>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const runExplain = async () => {
    setLoading(true)
    setError(null)
    setPlan(null)
    try {
      const result = await api.explain.run(connId, query, analyzeMode)
      setPlan(result.plan)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setLoading(false)
    }
  }

  const loadPasteJSON = () => {
    try {
      setPlan(JSON.parse(pasteJSON))
      setError(null)
    } catch {
      setError('Invalid JSON — paste the output of EXPLAIN (FORMAT JSON)')
    }
  }

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">EXPLAIN Plan</h1>

      <Tabs defaultValue={initialQuery ? 'direct' : 'direct'}>
        <TabsList>
          <TabsTrigger value="direct">Direct</TabsTrigger>
          <TabsTrigger value="paste">Paste JSON</TabsTrigger>
        </TabsList>

        <TabsContent value="direct" className="space-y-3">
          {analyzeMode && (
            <div className="rounded-md border border-yellow-400/50 bg-yellow-50 dark:bg-yellow-900/20 p-3 text-sm text-yellow-800 dark:text-yellow-200">
              Warning: EXPLAIN ANALYZE will execute the query on your database.
            </div>
          )}
          <Textarea
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            rows={5}
            className="font-mono text-sm"
            placeholder="SELECT ..."
          />
          <div className="flex items-center gap-4">
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={analyzeMode}
                onChange={(e) => setAnalyzeMode(e.target.checked)}
                className="rounded"
              />
              ANALYZE mode (executes query)
            </label>
            <Button onClick={runExplain} disabled={loading || !query.trim()}>
              {loading ? 'Running...' : 'Run EXPLAIN'}
            </Button>
          </div>
        </TabsContent>

        <TabsContent value="paste" className="space-y-3">
          <Textarea
            value={pasteJSON}
            onChange={(e) => setPasteJSON(e.target.value)}
            rows={8}
            className="font-mono text-sm"
            placeholder="Paste EXPLAIN (FORMAT JSON) output here..."
          />
          <Button onClick={loadPasteJSON} disabled={!pasteJSON.trim()}>
            Visualize
          </Button>
        </TabsContent>
      </Tabs>

      {error && (
        <div className="rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {plan != null && (
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-lg flex items-center gap-2">
              Execution Plan
              <Badge variant="outline">{analyzeMode ? 'ANALYZE' : 'ESTIMATE'}</Badge>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <ExplainJsonTree data={plan} />
          </CardContent>
        </Card>
      )}
    </div>
  )
}

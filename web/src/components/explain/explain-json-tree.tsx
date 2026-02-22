import { useState } from 'react'
import { cn } from '@/lib/utils'

interface PlanNode {
  'Node Type': string
  'Relation Name'?: string
  'Total Cost'?: number
  'Startup Cost'?: number
  'Actual Rows'?: number
  'Plan Rows'?: number
  'Actual Total Time'?: number
  'Shared Hit Blocks'?: number
  'Shared Read Blocks'?: number
  Plans?: PlanNode[]
  [key: string]: unknown
}

export function ExplainJsonTree({ data }: { data: unknown }) {
  // EXPLAIN FORMAT JSON returns an array with one element containing { Plan: ... }
  if (Array.isArray(data)) {
    const root = (data[0] as { Plan?: PlanNode })?.Plan
    if (!root) return <p className="text-muted-foreground">No plan data found.</p>
    return <PlanNodeView node={root} depth={0} />
  }
  // Direct plan node
  return <PlanNodeView node={data as PlanNode} depth={0} />
}

function PlanNodeView({ node, depth }: { node: PlanNode; depth: number }) {
  const [expanded, setExpanded] = useState(true)
  const hasChildren = node.Plans && node.Plans.length > 0
  const isExpensive = (node['Total Cost'] ?? 0) > 1000
  const isSeqScan = node['Node Type'] === 'Seq Scan'
  const rowMismatch =
    node['Plan Rows'] != null &&
    node['Actual Rows'] != null &&
    node['Actual Rows'] > node['Plan Rows'] * 10

  return (
    <div className={cn('relative', depth > 0 && 'ml-5 border-l border-border pl-3')}>
      <button
        type="button"
        className={cn(
          'flex flex-wrap items-center gap-x-2 py-1 cursor-pointer rounded px-1 hover:bg-muted/50 text-sm w-full text-left',
          isExpensive && 'text-destructive font-medium',
          isSeqScan && 'text-yellow-600 dark:text-yellow-400',
        )}
        onClick={() => setExpanded((e) => !e)}
      >
        <span className="text-muted-foreground">
          {hasChildren ? (expanded ? '\u25BC' : '\u25B6') : '\u25CB'}
        </span>
        <span className="font-semibold">{node['Node Type']}</span>
        {node['Relation Name'] && (
          <span className="text-muted-foreground">on {node['Relation Name']}</span>
        )}
        {node['Total Cost'] != null && (
          <span className="text-xs text-muted-foreground">
            cost={node['Startup Cost']?.toFixed(2)}..{node['Total Cost']?.toFixed(2)}
          </span>
        )}
        {node['Actual Total Time'] != null && (
          <span className="text-xs font-mono bg-muted px-1 rounded">
            {node['Actual Total Time']?.toFixed(2)}ms
          </span>
        )}
        {node['Actual Rows'] != null && (
          <span className="text-xs text-muted-foreground">rows={node['Actual Rows']}</span>
        )}
        {rowMismatch && (
          <span className="text-xs bg-destructive/10 text-destructive px-1 rounded">
            estimate off 10x+
          </span>
        )}
      </button>
      {expanded && hasChildren && (
        <div>
          {node.Plans?.map((child, i) => (
            <PlanNodeView key={i} node={child} depth={depth + 1} />
          ))}
        </div>
      )}
    </div>
  )
}

import type { PlanNode } from './explain-json-tree'

interface MySQLTable {
  table_name?: string
  access_type?: string
  key?: string
  possible_keys?: string[]
  rows_examined_per_scan?: number
  rows_produced_per_join?: number
  rows?: number
  filtered?: number
  cost_info?: { read_cost?: string; eval_cost?: string; query_cost?: string }
  attached_condition?: string
}

interface MySQLQueryBlock {
  select_id?: number
  cost_info?: { query_cost?: string }
  table?: MySQLTable
  nested_loop?: { table: MySQLTable }[]
  ordering_operation?: MySQLQueryBlock & { using_filesort?: boolean }
  grouping_operation?: MySQLQueryBlock
  duplicates_removal?: MySQLQueryBlock
  subqueries?: { query_block: MySQLQueryBlock }[]
}

const ACCESS_TYPE_LABELS: Record<string, string> = {
  ALL: 'Full Table Scan',
  index: 'Full Index Scan',
  range: 'Index Range Scan',
  ref: 'Index Lookup',
  eq_ref: 'Unique Index Lookup',
  const: 'Constant',
  system: 'System',
  fulltext: 'Fulltext Search',
  ref_or_null: 'Index Lookup (incl. NULL)',
  index_merge: 'Index Merge',
  unique_subquery: 'Unique Subquery',
  index_subquery: 'Index Subquery',
}

/** Parse MySQL EXPLAIN FORMAT=JSON into PlanNode tree. */
export function parseMySQLExplain(data: unknown): PlanNode | null {
  const qb = (data as { query_block?: MySQLQueryBlock })?.query_block
  if (!qb) return null
  return convertBlock(qb)
}

function convertBlock(block: MySQLQueryBlock): PlanNode {
  const children: PlanNode[] = []

  // Handle ordering_operation wrapper
  if (block.ordering_operation) {
    const child = convertBlock(block.ordering_operation)
    child['Node Type'] = `Sort -> ${child['Node Type']}`
    return child
  }

  // Handle grouping_operation wrapper
  if (block.grouping_operation) {
    const child = convertBlock(block.grouping_operation)
    child['Node Type'] = `Group -> ${child['Node Type']}`
    return child
  }

  // Handle nested_loop (multiple tables)
  if (block.nested_loop) {
    for (const nl of block.nested_loop) {
      children.push(convertTable(nl.table))
    }
    const cost = block.cost_info?.query_cost
      ? Number.parseFloat(block.cost_info.query_cost)
      : undefined
    return {
      'Node Type': 'Nested Loop',
      'Total Cost': cost,
      Plans: children,
    }
  }

  // Single table access
  if (block.table) {
    return convertTable(block.table)
  }

  // Fallback
  return {
    'Node Type': 'Query Block',
    'Total Cost': block.cost_info?.query_cost
      ? Number.parseFloat(block.cost_info.query_cost)
      : undefined,
    Plans: children.length > 0 ? children : undefined,
  }
}

function convertTable(table: MySQLTable): PlanNode {
  const accessType = table.access_type || 'unknown'
  const label = ACCESS_TYPE_LABELS[accessType] || accessType

  const node: PlanNode = {
    'Node Type': label,
    'Relation Name': table.table_name,
    'Plan Rows': table.rows_examined_per_scan ?? table.rows,
  }

  if (table.key) {
    node['Index Name'] = table.key
  }
  if (table.filtered != null) {
    node.Filter = `${table.filtered}%`
  }
  if (table.cost_info?.read_cost) {
    node['Total Cost'] = Number.parseFloat(table.cost_info.read_cost)
  }
  if (table.attached_condition) {
    node['Filter Condition'] = table.attached_condition
  }
  return node
}

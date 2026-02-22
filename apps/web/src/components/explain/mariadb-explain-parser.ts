import type { PlanNode } from './explain-json-tree'

interface MariaDBTable {
  table_name?: string
  access_type?: string
  key?: string
  possible_keys?: string[]
  rows?: number
  filtered?: number
  r_rows?: number
  r_filtered?: number
  r_total_time_ms?: number | string
  attached_condition?: string
}

interface MariaDBQueryBlock {
  select_id?: number
  table?: MariaDBTable
  nested_loop?: { table: MariaDBTable }[]
  ordering_operation?: MariaDBQueryBlock & { using_filesort?: boolean }
  grouping_operation?: MariaDBQueryBlock
  read_sorted_file?: MariaDBQueryBlock
  subqueries?: { query_block: MariaDBQueryBlock }[]
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
}

/** Parse MariaDB EXPLAIN/ANALYZE FORMAT=JSON into PlanNode tree. */
export function parseMariaDBExplain(data: unknown): PlanNode | null {
  const qb = (data as { query_block?: MariaDBQueryBlock })?.query_block
  if (!qb) return null
  return convertBlock(qb)
}

function convertBlock(block: MariaDBQueryBlock): PlanNode {
  const children: PlanNode[] = []

  if (block.ordering_operation) {
    const child = convertBlock(block.ordering_operation)
    child['Node Type'] = `Sort -> ${child['Node Type']}`
    return child
  }

  if (block.grouping_operation) {
    const child = convertBlock(block.grouping_operation)
    child['Node Type'] = `Group -> ${child['Node Type']}`
    return child
  }

  if (block.read_sorted_file) {
    const child = convertBlock(block.read_sorted_file)
    child['Node Type'] = `Filesort -> ${child['Node Type']}`
    return child
  }

  if (block.nested_loop) {
    for (const nl of block.nested_loop) {
      children.push(convertTable(nl.table))
    }
    return { 'Node Type': 'Nested Loop', Plans: children }
  }

  if (block.table) {
    return convertTable(block.table)
  }

  return { 'Node Type': 'Query Block', Plans: children.length > 0 ? children : undefined }
}

function convertTable(table: MariaDBTable): PlanNode {
  const accessType = table.access_type || 'unknown'
  const label = ACCESS_TYPE_LABELS[accessType] || accessType

  const node: PlanNode = {
    'Node Type': label,
    'Relation Name': table.table_name,
    'Plan Rows': table.rows,
  }

  if (table.key) {
    node['Index Name'] = table.key
  }
  if (table.filtered != null) {
    node.Filter = `${table.filtered}%`
  }

  // ANALYZE mode fields (actual execution stats)
  if (table.r_rows != null) {
    node['Actual Rows'] = table.r_rows
  }
  if (table.r_total_time_ms != null) {
    node['Actual Total Time'] =
      typeof table.r_total_time_ms === 'string'
        ? Number.parseFloat(table.r_total_time_ms)
        : table.r_total_time_ms
  }
  if (table.attached_condition) {
    node['Filter Condition'] = table.attached_condition
  }
  return node
}

export interface Connection {
  id: number
  name: string
  db_type: string
  created_at: string
  updated_at: string
}

export interface SlowQuery {
  query_id: string
  query: string
  calls: number
  total_exec_ms: number
  mean_exec_ms: number
  rows: number
  snapshot_at: string
}

export interface QueryDelta extends SlowQuery {
  calls_delta: number
  total_exec_delta_ms: number
  mean_exec_delta_ms: number
  period_secs: number
}

export interface QuerySnapshot {
  id: number
  connection_id: number
  queries: SlowQuery[]
  captured_at: string
}

export interface IndexStat {
  schema_name: string
  table_name: string
  index_name: string
  index_scans: number
  tup_read: number
  tup_fetch: number
  index_size_bytes: number
  is_unused: boolean
}

export interface ExplainPlan {
  query: string
  plan: unknown
}

export interface DatabaseStats {
  db_name: string
  size_bytes: number
  active_conns: number
  max_conns: number
  cache_hit_ratio: number
}

export interface Recommendation {
  type: string
  schema_name: string
  table_name: string
  index_name?: string
  description: string
  sql?: string
  severity: 'high' | 'medium' | 'low'
}

export interface DuplicateIndex {
  table_name: string
  index1: string
  index2: string
  index_def: string
}

export interface IndexAnalysisResult {
  unused_indexes: IndexStat[]
  missing_candidates: TableStat[]
  duplicate_indexes: DuplicateIndex[]
  recommendations: Recommendation[]
  captured_at: string
}

export interface TableStat {
  schema_name: string
  table_name: string
  seq_scans: number
  seq_tup_read: number
  n_live_tup: number
  table_size_bytes: number
}

export interface ApiError {
  error: string
}

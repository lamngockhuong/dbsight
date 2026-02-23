import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  type SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { QueryDelta } from '@/types'

const col = createColumnHelper<QueryDelta>()

interface SlowQueryTableProps {
  data: QueryDelta[]
  onSelect: (q: QueryDelta) => void
}

export function SlowQueryTable({ data, onSelect }: SlowQueryTableProps) {
  const [sorting, setSorting] = useState<SortingState>([{ id: 'total_exec_ms', desc: true }])
  const [globalFilter, setGlobalFilter] = useState('')

  const columns = [
    col.display({ id: 'rank', header: '#', cell: (info) => info.row.index + 1 }),
    col.accessor('query', {
      header: 'Query',
      cell: (info) => (
        <code className="text-xs max-w-md block truncate" title={info.getValue()}>
          {info.getValue().slice(0, 120)}
        </code>
      ),
      enableSorting: false,
    }),
    col.accessor('calls', { header: 'Calls' }),
    col.accessor('mean_exec_ms', {
      header: 'Avg (ms)',
      cell: (info) => info.getValue().toFixed(2),
    }),
    col.accessor('total_exec_ms', {
      header: 'Total (ms)',
      cell: (info) => info.getValue().toFixed(2),
    }),
    col.accessor('calls_delta', { header: 'Δ Calls' }),
    col.display({
      id: 'actions',
      header: '',
      cell: (info) => (
        <Button size="sm" variant="ghost" onClick={() => onSelect(info.row.original)}>
          Detail
        </Button>
      ),
    }),
  ]

  const table = useReactTable({
    data,
    columns,
    state: { sorting, globalFilter },
    onSortingChange: setSorting,
    onGlobalFilterChange: setGlobalFilter,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  })

  return (
    <div className="space-y-4">
      <Input
        placeholder="Filter queries..."
        value={globalFilter}
        onChange={(e) => setGlobalFilter(e.target.value)}
        className="max-w-sm"
      />
      <div className="overflow-x-auto rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((hg) => (
              <TableRow key={hg.id}>
                {hg.headers.map((h) => (
                  <TableHead
                    key={h.id}
                    onClick={h.column.getToggleSortingHandler()}
                    className={h.column.getCanSort() ? 'cursor-pointer select-none' : ''}
                  >
                    {flexRender(h.column.columnDef.header, h.getContext())}
                    {h.column.getIsSorted() === 'asc'
                      ? ' ↑'
                      : h.column.getIsSorted() === 'desc'
                        ? ' ↓'
                        : ''}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={columns.length} className="text-center text-muted-foreground">
                  No queries found
                </TableCell>
              </TableRow>
            ) : (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}

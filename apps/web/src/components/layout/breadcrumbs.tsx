import { Link, useLocation } from 'react-router-dom'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'

const routeLabels: Record<string, string> = {
  connections: 'Connections',
  queries: 'Slow Queries',
  explain: 'EXPLAIN',
  indexes: 'Indexes',
  paste: 'Paste Mode',
}

interface BreadcrumbsProps {
  connectionId?: number
  connectionName?: string
}

export function Breadcrumbs({ connectionId, connectionName }: BreadcrumbsProps) {
  const { pathname } = useLocation()

  const segments = pathname.split('/').filter(Boolean)
  if (segments.length === 0) return null

  const pageKey = segments[0]
  const pageLabel = routeLabels[pageKey]
  if (!pageLabel) return null

  const connLabel = connectionId ? connectionName || `Connection #${connectionId}` : null

  const items: { label: string; href?: string }[] = [{ label: 'Dashboard', href: '/' }]

  if (connLabel && connectionId) {
    items.push({ label: connLabel, href: `/queries/${connectionId}` })
  }

  items.push({ label: pageLabel })

  return (
    <Breadcrumb className="mb-4">
      <BreadcrumbList>
        {items.map((item, i) => {
          const isLast = i === items.length - 1
          return (
            <BreadcrumbItem key={item.label}>
              {i > 0 && <BreadcrumbSeparator />}
              {isLast ? (
                <BreadcrumbPage>{item.label}</BreadcrumbPage>
              ) : (
                <BreadcrumbLink asChild>
                  <Link
                    to={item.href ?? '/'}
                    className={connLabel ? 'max-w-[150px] truncate sm:max-w-none' : ''}
                  >
                    {item.label}
                  </Link>
                </BreadcrumbLink>
              )}
            </BreadcrumbItem>
          )
        })}
      </BreadcrumbList>
    </Breadcrumb>
  )
}

import { BarChart3, ClipboardPaste, Database, FileSearch, Home, Search } from 'lucide-react'
import { NavLink } from 'react-router-dom'
import { cn } from '@/lib/utils'

const baseNav = [
  { to: '/', label: 'Dashboard', icon: Home },
  { to: '/connections', label: 'Connections', icon: Database },
  { to: '/paste', label: 'Paste Mode', icon: ClipboardPaste },
]

interface SidebarNavProps {
  connectionId?: number
  onNavigate?: () => void
}

export function SidebarNav({ connectionId, onNavigate }: SidebarNavProps) {
  const connNav = connectionId
    ? [
        { to: `/queries/${connectionId}`, label: 'Slow Queries', icon: Search },
        { to: `/explain/${connectionId}`, label: 'EXPLAIN', icon: FileSearch },
        { to: `/indexes/${connectionId}`, label: 'Indexes', icon: BarChart3 },
      ]
    : []

  return (
    <nav className="flex flex-col gap-1">
      {[...baseNav, ...connNav].map((item) => (
        <NavLink
          key={item.to}
          to={item.to}
          onClick={onNavigate}
          className={({ isActive }) =>
            cn(
              'flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors',
              isActive
                ? 'bg-primary text-primary-foreground'
                : 'text-muted-foreground hover:bg-muted hover:text-foreground',
            )
          }
        >
          <item.icon className="h-4 w-4" />
          {item.label}
        </NavLink>
      ))}
    </nav>
  )
}

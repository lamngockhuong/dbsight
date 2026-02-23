import { ThemeToggle } from '@/components/theme/theme-toggle'
import { SidebarFooter } from './sidebar-footer'
import { SidebarNav } from './sidebar-nav'

interface SidebarProps {
  connectionId?: number
}

export function Sidebar({ connectionId }: SidebarProps) {
  return (
    <aside className="hidden md:flex w-56 border-r bg-muted/40 p-4 flex-col justify-between">
      <div>
        <div className="flex items-center justify-between mb-4 px-2">
          <h2 className="text-lg font-bold">DBSight</h2>
          <ThemeToggle />
        </div>
        <SidebarNav connectionId={connectionId} />
      </div>
      <SidebarFooter />
    </aside>
  )
}

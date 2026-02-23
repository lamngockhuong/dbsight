import { Breadcrumbs } from './breadcrumbs'
import { MobileSidebar } from './mobile-sidebar'
import { Sidebar } from './sidebar'

interface LayoutProps {
  children: React.ReactNode
  connectionId?: number
  connectionName?: string
}

export function Layout({ children, connectionId, connectionName }: LayoutProps) {
  return (
    <div className="flex min-h-screen bg-background">
      <Sidebar connectionId={connectionId} />
      <div className="flex-1 flex flex-col">
        <div className="md:hidden sticky top-0 z-40 flex items-center border-b bg-background px-3 py-2">
          <MobileSidebar connectionId={connectionId} />
          <span className="ml-2 font-bold">DBSight</span>
        </div>
        <main className="flex-1 p-3 md:p-6 overflow-auto">
          <Breadcrumbs connectionId={connectionId} connectionName={connectionName} />
          {children}
        </main>
      </div>
    </div>
  )
}

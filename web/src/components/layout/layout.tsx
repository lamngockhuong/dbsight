import { Sidebar } from './sidebar'

interface LayoutProps {
  children: React.ReactNode
  connectionId?: number
}

export function Layout({ children, connectionId }: LayoutProps) {
  return (
    <div className="flex min-h-screen bg-background">
      <Sidebar connectionId={connectionId} />
      <main className="flex-1 p-6 overflow-auto">{children}</main>
    </div>
  )
}

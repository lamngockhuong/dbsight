import { Menu } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useLocation } from 'react-router-dom'
import { ThemeToggle } from '@/components/theme/theme-toggle'
import { Button } from '@/components/ui/button'
import { Sheet, SheetContent, SheetTitle, SheetTrigger } from '@/components/ui/sheet'
import { SidebarFooter } from './sidebar-footer'
import { SidebarNav } from './sidebar-nav'

interface MobileSidebarProps {
  connectionId?: number
}

export function MobileSidebar({ connectionId }: MobileSidebarProps) {
  const [open, setOpen] = useState(false)
  const { pathname } = useLocation()

  // biome-ignore lint/correctness/useExhaustiveDependencies: close sheet on route change
  useEffect(() => {
    setOpen(false)
  }, [pathname])

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="ghost" size="icon" className="md:hidden">
          <Menu className="h-5 w-5" />
          <span className="sr-only">Toggle menu</span>
        </Button>
      </SheetTrigger>
      <SheetContent side="left" className="w-64 p-4 flex flex-col justify-between">
        <div>
          <div className="flex items-center justify-between mb-4 px-2">
            <SheetTitle className="text-lg font-bold">DBSight</SheetTitle>
            <ThemeToggle />
          </div>
          <SidebarNav connectionId={connectionId} onNavigate={() => setOpen(false)} />
        </div>
        <SidebarFooter />
      </SheetContent>
    </Sheet>
  )
}

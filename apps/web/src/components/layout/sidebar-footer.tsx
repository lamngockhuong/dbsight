import { ExternalLink, Github } from 'lucide-react'

export function SidebarFooter() {
  return (
    <div className="border-t pt-3 space-y-1">
      <p className="text-xs text-muted-foreground">DBSight v{__APP_VERSION__}</p>
      <p className="text-xs text-muted-foreground">Built by Khuong Lam</p>
      <div className="flex items-center gap-2">
        <a
          href="https://github.com/lamngockhuong/dbsight"
          target="_blank"
          rel="noopener noreferrer"
          className="text-muted-foreground hover:text-foreground transition-colors"
        >
          <Github className="h-4 w-4" />
        </a>
        <a
          href="https://dbsight.khuong.dev"
          target="_blank"
          rel="noopener noreferrer"
          className="text-muted-foreground hover:text-foreground transition-colors"
        >
          <ExternalLink className="h-4 w-4" />
        </a>
      </div>
    </div>
  )
}

import { Link } from '@tanstack/react-router'
import { CompassIcon, FileTextIcon } from 'lucide-react'

import { Button } from '#/components/ui/button'
import ThemeToggle from './ThemeToggle'

export default function Header() {
  return (
    <header className="sticky top-0 z-50 border-b border-border/70 bg-background/80 px-4 backdrop-blur-xl">
      <nav className="page-wrap flex flex-wrap items-center gap-3 py-3 sm:py-4">
        <h2 className="m-0 flex-shrink-0 text-base font-semibold tracking-tight">
          <Link
            to="/"
            className="inline-flex items-center gap-3 rounded-full border border-border/70 bg-card/85 px-4 py-2 text-sm text-foreground no-underline shadow-[0_16px_40px_rgba(15,23,42,0.06)]"
          >
            <span className="inline-flex size-2.5 rounded-full bg-[linear-gradient(90deg,#56c6be,#7ed3bf)]" />
            <span className="font-semibold">EasyDrop</span>
          </Link>
        </h2>

        <div className="ml-auto flex items-center gap-2">
          <Button asChild size="sm" variant="ghost">
            <Link to="/">
              <CompassIcon data-icon="inline-start" />
              时间流
            </Link>
          </Button>
          <Button asChild size="sm" variant="ghost">
            <Link to="/about">
              <FileTextIcon data-icon="inline-start" />
              关于
            </Link>
          </Button>
          <ThemeToggle />
        </div>
      </nav>
    </header>
  )
}

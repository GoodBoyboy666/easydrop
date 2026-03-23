export function SiteFooter() {
  return (
    <footer className="border-t border-border/70 bg-background/90">
      <div className="mx-auto flex w-full max-w-7xl flex-col gap-2 px-4 py-6 text-sm text-muted-foreground sm:px-6 lg:px-8 md:flex-row md:items-center md:justify-between">
        <div>一言预留</div>
        <div>© {new Date().getFullYear()} EasyDrop</div>
      </div>
    </footer>
  )
}

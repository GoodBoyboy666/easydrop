export default function Footer() {
  const year = new Date().getFullYear()

  return (
    <footer className="mt-20 border-t border-border/70 px-4 pb-14 pt-8 text-muted-foreground">
      <div className="page-wrap flex flex-col gap-3 text-sm sm:flex-row sm:items-center sm:justify-between">
        <p className="m-0">
          &copy; {year} EasyDrop.
        </p>
        <p className="island-kicker m-0">
          Markdown only · No executable script
        </p>
      </div>
    </footer>
  )
}

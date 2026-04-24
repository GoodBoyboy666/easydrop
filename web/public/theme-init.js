;(() => {
    try {
        const storageKey = 'easydrop-theme'
        const storedTheme = window.localStorage.getItem(storageKey)
        const themeMode =
            storedTheme === 'light' ||
            storedTheme === 'dark' ||
            storedTheme === 'system'
                ? storedTheme
                : 'system'
        const resolvedTheme =
            themeMode === 'light' || themeMode === 'dark'
                ? themeMode
                : window.matchMedia('(prefers-color-scheme: dark)').matches
                    ? 'dark'
                    : 'light'

        document.documentElement.dataset.themeMode = themeMode
        document.documentElement.classList.toggle(
            'dark',
            resolvedTheme === 'dark',
        )
        document.documentElement.style.colorScheme = resolvedTheme
    } catch {
        const fallbackTheme = window.matchMedia(
            '(prefers-color-scheme: dark)',
        ).matches
            ? 'dark'
            : 'light'

        document.documentElement.dataset.themeMode = 'system'
        document.documentElement.classList.toggle(
            'dark',
            fallbackTheme === 'dark',
        )
        document.documentElement.style.colorScheme = fallbackTheme
    }
})()
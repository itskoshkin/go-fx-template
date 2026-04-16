(function () { /* Manage persisted UI theme selection and expose helpers globally. */
    const STORAGE_KEY = 'go_fx_template.theme'; /* Keep the persisted theme key stable across the app. */
    const THEMES = { /* Enumerate all supported theme modes. */
        light: 'light', /* Force the light palette. */
        system: 'system', /* Follow the operating-system preference. */
        dark: 'dark', /* Force the dark palette. */
    };

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)'); /* Read the operating-system color-scheme preference. */
    let currentTheme = window.localStorage.getItem(STORAGE_KEY) || THEMES.system; /* Restore the persisted theme or fall back to system mode. */

    if (!Object.prototype.hasOwnProperty.call(THEMES, currentTheme)) { /* Guard against invalid persisted theme values. */
        currentTheme = THEMES.system; /* Fall back to system mode when localStorage holds an invalid value. */
    }

    function getResolvedTheme(theme) { /* Resolve a requested theme mode into an actual light or dark palette. */
        if (theme === THEMES.system) { /* Handle system-driven theme resolution explicitly. */
            return mediaQuery.matches ? THEMES.dark : THEMES.light; /* Follow the current operating-system preference. */
        }

        return theme === THEMES.light ? THEMES.light : THEMES.dark; /* Normalize all non-system values to valid concrete themes. */
    }

    function syncThemeColorMeta(resolvedTheme) { /* Keep the browser theme-color meta tag aligned with the active palette. */
        const meta = document.querySelector('meta[name="theme-color"]'); /* Find the browser theme-color meta tag when it exists. */

        if (!meta) { /* Guard against pages that omit the theme-color tag. */
            return; /* Stop before mutating a missing node. */
        }

        meta.setAttribute('content', resolvedTheme === THEMES.light ? '#f4f1ea' : '#000000'); /* Match browser chrome to the active palette. */
    }

    function applyTheme(theme, notify) { /* Apply one requested theme mode to the document root. */
        const resolvedTheme = getResolvedTheme(theme); /* Resolve the requested theme into a concrete light or dark palette. */
        const root = document.documentElement; /* Use the document root as the single theme-class host. */

        root.classList.remove('theme-light', 'theme-dark'); /* Remove any previously applied concrete theme classes. */
        root.classList.add(`theme-${resolvedTheme}`); /* Apply the resolved concrete theme class. */
        root.dataset.themePreference = theme; /* Store the requested theme mode for CSS or debugging when needed. */
        syncThemeColorMeta(resolvedTheme); /* Keep the browser theme-color meta tag aligned with the active palette. */

        if (notify) { /* Emit a change event only for externally visible theme changes. */
            document.dispatchEvent(new CustomEvent('app:themechange', {detail: {theme: theme, resolvedTheme: resolvedTheme}})); /* Notify page scripts about the new theme selection. */
        }
    }

    function setTheme(theme) { /* Persist and apply one requested theme mode. */
        if (!Object.prototype.hasOwnProperty.call(THEMES, theme)) { /* Guard against invalid theme requests. */
            return; /* Stop before mutating persisted theme state. */
        }

        currentTheme = theme; /* Store the requested theme mode in memory. */
        window.localStorage.setItem(STORAGE_KEY, theme); /* Persist the requested theme mode to localStorage. */
        applyTheme(theme, true); /* Apply the new theme and notify page scripts. */
    }

    mediaQuery.addEventListener('change', () => { /* React to operating-system theme changes while system mode is active. */
        if (currentTheme === THEMES.system) { /* Re-apply the theme only when system mode currently drives the palette. */
            applyTheme(THEMES.system, true); /* Refresh the resolved theme and notify listeners. */
        }
    });

    window.AppTheme = { /* Expose theme helpers globally for page scripts. */
        THEMES: THEMES, /* Share the supported theme modes. */
        getTheme: () => currentTheme, /* Share the currently requested theme mode. */
        setTheme: setTheme, /* Share the persisted theme setter. */
    };

    applyTheme(currentTheme, false); /* Apply the persisted theme immediately during head parsing. */
}());

document.addEventListener('DOMContentLoaded', () => { /* Wait for the DOM before wiring global preference controls. */
    const languageButton = document.getElementById('open-language-modal-btn'); /* Cache the fixed language FAB. */
    const themeButton = document.getElementById('open-theme-modal-btn'); /* Cache the fixed theme FAB. */
    const languageModal = document.getElementById('language-modal'); /* Cache the language-selection dialog. */
    const themeModal = document.getElementById('theme-modal'); /* Cache the theme-selection dialog. */

    if (!languageButton || !themeButton || !languageModal || !themeModal || !window.AppI18n || !window.AppTheme) { /* Guard against pages that do not render the global preference controls. */
        return; /* Stop before wiring missing preference nodes. */
    }

    function showDialog(dialog) { /* Open one preference dialog safely across browsers. */
        if (dialog.open || dialog.hasAttribute('open')) { /* Ignore duplicate open requests for the same already visible dialog. */
            return; /* Stop before the native dialog API throws on repeated showModal calls. */
        }

        dialog.dataset.dialogFallback = 'true'; /* Force compact preference popovers to use the non-modal fallback path. */
        dialog.setAttribute('open', 'open'); /* Fall back to the open attribute when showModal is unavailable. */
    }

    function closeDialog(dialog) { /* Close one preference dialog safely across browsers. */
        if (!dialog.open && !dialog.hasAttribute('open')) { /* Ignore duplicate close requests for dialogs that are already closed. */
            return; /* Stop before the native dialog API throws on repeated close calls. */
        }

        dialog.removeAttribute('open'); /* Fall back to removing the open attribute when close is unavailable. */
        dialog.dispatchEvent(new Event('close')); /* Emit a synthetic close event so future cleanup hooks still work. */
    }

    function closeAllDialogs() { /* Close both preference dialogs before opening another one. */
        closeDialog(languageModal); /* Close the language-selection dialog. */
        closeDialog(themeModal); /* Close the theme-selection dialog. */
    }

    function isDialogOpen(dialog) { /* Detect dialog visibility consistently for compact preference popovers. */
        return dialog.open || dialog.hasAttribute('open'); /* Treat either native or attribute-driven open state as visible. */
    }

    function syncLanguageUI() { /* Refresh the current-language label and active state markers. */
        const language = window.AppI18n.getLanguage(); /* Read the current active language once. */

        document.querySelectorAll('[data-language-option]').forEach((button) => { /* Refresh active state for every language option button. */
            button.classList.toggle('is-active', button.dataset.languageOption === language); /* Highlight the currently selected language option. */
        });
    }

    function syncThemeUI() { /* Refresh the current-theme label and active state markers. */
        const theme = window.AppTheme.getTheme(); /* Read the currently requested theme mode once. */

        document.querySelectorAll('[data-theme-option]').forEach((button) => { /* Refresh active state for every theme option button. */
            button.classList.toggle('is-active', button.dataset.themeOption === theme); /* Highlight the currently selected theme option. */
        });
    }

    languageButton.addEventListener('click', (event) => { /* Toggle the language-selection dialog from the fixed FAB. */
        event.stopPropagation(); /* Keep the global outside-click handler from immediately closing the dialog again. */
        if (isDialogOpen(languageModal)) { /* Treat repeated clicks on the same FAB as a toggle close action. */
            closeDialog(languageModal); /* Close the currently visible language popover. */
            return; /* Stop before reopening it. */
        }

        closeDialog(themeModal); /* Ensure only one preference dialog stays open at a time. */
        syncLanguageUI(); /* Refresh current language markers immediately before opening the dialog. */
        showDialog(languageModal); /* Open the language dialog. */
    });

    themeButton.addEventListener('click', (event) => { /* Toggle the theme-selection dialog from the fixed FAB. */
        event.stopPropagation(); /* Keep the global outside-click handler from immediately closing the dialog again. */
        if (isDialogOpen(themeModal)) { /* Treat repeated clicks on the same FAB as a toggle close action. */
            closeDialog(themeModal); /* Close the currently visible theme popover. */
            return; /* Stop before reopening it. */
        }

        closeDialog(languageModal); /* Ensure only one preference dialog stays open at a time. */
        syncThemeUI(); /* Refresh current theme markers immediately before opening the dialog. */
        showDialog(themeModal); /* Open the theme dialog. */
    });

    document.querySelectorAll('[data-language-option]').forEach((button) => { /* Bind every language option button. */
        button.addEventListener('click', (event) => { /* Apply the requested language from the clicked option. */
            event.stopPropagation(); /* Prevent the global outside-click handler from racing with the selection click. */
            window.AppI18n.setLanguage(button.dataset.languageOption); /* Persist and apply the requested language. */
            closeDialog(languageModal); /* Close the language dialog immediately after selection. */
        });
    });

    document.querySelectorAll('[data-theme-option]').forEach((button) => { /* Bind every theme option button. */
        button.addEventListener('click', (event) => { /* Apply the requested theme from the clicked option. */
            event.stopPropagation(); /* Prevent the global outside-click handler from racing with the selection click. */
            window.AppTheme.setTheme(button.dataset.themeOption); /* Persist and apply the requested theme. */
            closeDialog(themeModal); /* Close the theme dialog immediately after selection. */
        });
    });

    languageModal.addEventListener('click', (event) => { /* Stop inner popover clicks from bubbling into the global outside-click handler. */
        event.stopPropagation(); /* Keep clicks inside the language popover from closing it. */
    });

    themeModal.addEventListener('click', (event) => { /* Stop inner popover clicks from bubbling into the global outside-click handler. */
        event.stopPropagation(); /* Keep clicks inside the theme popover from closing it. */
    });

    document.addEventListener('click', (event) => { /* Close popovers when the user clicks anywhere outside both FABs and popovers. */
        const target = event.target; /* Read the clicked node once for the outside-click checks. */

        if (target instanceof Node && (languageButton.contains(target) || themeButton.contains(target) || languageModal.contains(target) || themeModal.contains(target))) { /* Ignore clicks that originate inside any FAB or preference popover. */
            return; /* Stop before closing currently open preference popovers. */
        }

        closeAllDialogs(); /* Close any visible preference popovers after an outside click. */
    });

    document.addEventListener('app:languagechange', syncLanguageUI); /* Keep the language FAB and option states aligned after language changes. */
    document.addEventListener('app:themechange', syncThemeUI); /* Keep the theme FAB and option states aligned after theme changes. */

    syncLanguageUI(); /* Render the initial language FAB state. */
    syncThemeUI(); /* Render the initial theme FAB state. */
});

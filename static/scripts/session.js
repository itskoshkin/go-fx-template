(function () { /* Bootstrap auth-cookie mirroring and route guards before the page renders. */
    const APP_STORAGE_KEYS = { /* Keep localStorage auth keys in one place. */
        accessToken: 'go_fx_template.access_token', /* Store the access token under a stable key. */
    };

    const APP_COOKIE_KEYS = { /* Keep browser cookie auth keys in one place. */
        accessToken: 'go_fx_template.access_token', /* Mirror the access token into a readable cookie. */
        sessionMarker: 'go_fx_template.session', /* Keep a lightweight readable session marker for SSR guards. */
    };

    const TOKEN_COOKIE_MAX_AGE_SECONDS = 31536000; /* Match the long-lived browser session horizon used elsewhere. */

    function readCookie(name) { /* Read one cookie value by name from document.cookie. */
        const escapedName = name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'); /* Escape special characters before building the lookup pattern. */
        const match = document.cookie.match(new RegExp(`(?:^|; )${escapedName}=([^;]*)`)); /* Find the matching cookie in the flat cookie string. */
        return match ? decodeURIComponent(match[1]) : ''; /* Decode the cookie value or fall back to empty string. */
    }

    function writeCookie(name, value) { /* Persist one readable auth cookie with a stable path and lifetime. */
        document.cookie = `${name}=${encodeURIComponent(value)}; Max-Age=${TOKEN_COOKIE_MAX_AGE_SECONDS}; Path=/; SameSite=Lax`; /* Write the encoded cookie into the browser cookie jar. */
    }

    function clearCookie(name) { /* Remove one readable auth cookie by name. */
        document.cookie = `${name}=; Max-Age=0; Path=/; SameSite=Lax`; /* Expire the cookie immediately across the whole site. */
    }

    function currentPath() { /* Read the current route path without query-string noise. */
        return window.location.pathname || '/'; /* Fall back to the root route when the browser returns an empty path. */
    }

    function bootstrapSessionGuard() { /* Mirror the access token into a cookie and redirect away from invalid routes early. */
        const path = currentPath(); /* Read the current route path once for branch checks. */
        const isLoginPage = path === '/login'; /* Detect the public auth page. */
        const requiresSession = path === '/' || path === '/change-password' || path === '/delete-account'; /* Detect routes that require an authenticated browser session. */

        if (!isLoginPage && !requiresSession) { /* Skip work on routes that do not need early auth guarding. */
            return; /* Stop before reading browser storage unnecessarily. */
        }

        try { /* Read current browser auth state and normalize it for SSR route guards. */
            const accessToken = window.localStorage.getItem(APP_STORAGE_KEYS.accessToken) || ''; /* Read the stored access token from localStorage when present. */
            let accessTokenCookie = readCookie(APP_COOKIE_KEYS.accessToken); /* Read the mirrored access-token cookie when present. */
            const sessionMarkerCookie = readCookie(APP_COOKIE_KEYS.sessionMarker); /* Read the lightweight readable session marker cookie when present. */

            if (accessTokenCookie === '' && accessToken !== '') { /* Mirror the access token into a readable cookie when only localStorage has it. */
                writeCookie(APP_COOKIE_KEYS.accessToken, accessToken); /* Keep SSR route checks aligned with the stored access token. */
                accessTokenCookie = accessToken; /* Treat the just-written cookie as present for the current guard run. */
            }

            const hasSession = accessToken !== '' || accessTokenCookie !== '' || sessionMarkerCookie !== ''; /* Treat any readable browser auth marker as an active session. */

            if (isLoginPage && hasSession) { /* Prevent authenticated users from lingering on the login page. */
                window.location.replace('/'); /* Redirect authenticated users to the dashboard immediately. */
                return; /* Stop after scheduling the redirect. */
            }

            if (requiresSession && !hasSession) { /* Prevent unauthenticated users from opening protected browser routes. */
                window.location.replace('/login'); /* Redirect unauthenticated users to the login page immediately. */
            }
        } catch (_error) { /* Handle localStorage access failures defensively. */
            if (requiresSession) { /* Protect authenticated-only routes even when storage access fails. */
                window.location.replace('/login'); /* Fall back to the login page when browser storage cannot be read safely. */
            }
        }
    }

    window.AppSessionBootstrap = { /* Expose shared auth-cookie helpers to later deferred scripts. */
        APP_STORAGE_KEYS: APP_STORAGE_KEYS, /* Share localStorage auth-key names. */
        APP_COOKIE_KEYS: APP_COOKIE_KEYS, /* Share cookie auth-key names. */
        TOKEN_COOKIE_MAX_AGE_SECONDS: TOKEN_COOKIE_MAX_AGE_SECONDS, /* Share the stable cookie lifetime constant. */
        readCookie: readCookie, /* Share the cookie reader helper. */
        writeCookie: writeCookie, /* Share the cookie writer helper. */
        clearCookie: clearCookie, /* Share the cookie clearer helper. */
    };

    bootstrapSessionGuard(); /* Run the early session guard immediately during head parsing. */
}());

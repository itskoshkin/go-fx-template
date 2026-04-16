const API_BASE = '/api/v1'; /* Keep the API prefix in one place for all frontend requests. */
const SESSION_BOOTSTRAP = window.AppSessionBootstrap || {}; /* Reuse cookie and storage helpers exposed by the early session bootstrap when available. */

const APP_STORAGE_KEYS = SESSION_BOOTSTRAP.APP_STORAGE_KEYS || { /* Collect all localStorage keys in one place. */
    accessToken: 'go_fx_template.access_token', /* Store the access token under a stable key. */
};

const APP_COOKIE_KEYS = SESSION_BOOTSTRAP.APP_COOKIE_KEYS || { /* Collect all auth cookie keys in one place. */
    accessToken: 'go_fx_template.access_token', /* Mirror the access token into a browser cookie. */
    sessionMarker: 'go_fx_template.session', /* Mirror a lightweight session marker cookie for SSR route gating. */
};

const TOKEN_COOKIE_MAX_AGE_SECONDS = SESSION_BOOTSTRAP.TOKEN_COOKIE_MAX_AGE_SECONDS || 31536000; /* Keep token cookies alive long enough to match persistent localStorage sessions. */

function goHome() { /* Navigate back to the dashboard page. */
    window.location.href = '/'; /* Replace the current location with the dashboard route. */
}

function goHomeToProfile() { /* Navigate back to the dashboard and request reopening the profile modal. */
    window.location.href = '/?modal=profile'; /* Replace the current location with the dashboard route and a profile-modal hint. */
}

function goTo(path) { /* Navigate to an arbitrary relative route. */
    window.location.href = path; /* Replace the current location with the provided path. */
}

function getPageRoot() { /* Return the document body for data attribute lookups. */
    return document.body; /* Use the body element as the single source of page metadata. */
}

function getCurrentPageName() { /* Read the logical page name from the body. */
    return getPageRoot()?.dataset.page || ''; /* Fall back to an empty string when the marker is missing. */
}

function readCookie(name) { /* Read one cookie value by name from document.cookie. */
    if (typeof SESSION_BOOTSTRAP.readCookie === 'function') { /* Reuse the early bootstrap cookie reader when it is available. */
        return SESSION_BOOTSTRAP.readCookie(name); /* Delegate to the shared cookie reader helper. */
    }

    const escapedName = name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'); /* Escape special characters before building the cookie lookup pattern. */
    const match = document.cookie.match(new RegExp(`(?:^|; )${escapedName}=([^;]*)`)); /* Find the matching cookie in the flat cookie string. */
    return match ? decodeURIComponent(match[1]) : ''; /* Decode the cookie value or fall back to empty string. */
}

function writeCookie(name, value) { /* Persist one cookie value with a stable path and lifetime. */
    if (typeof SESSION_BOOTSTRAP.writeCookie === 'function') { /* Reuse the early bootstrap cookie writer when it is available. */
        SESSION_BOOTSTRAP.writeCookie(name, value); /* Delegate to the shared cookie writer helper. */
        return; /* Stop after the shared writer has persisted the cookie. */
    }

    document.cookie = `${name}=${encodeURIComponent(value)}; Max-Age=${TOKEN_COOKIE_MAX_AGE_SECONDS}; Path=/; SameSite=Lax`; /* Write the encoded cookie into the browser cookie jar. */
}

function clearCookie(name) { /* Remove one cookie value by name. */
    if (typeof SESSION_BOOTSTRAP.clearCookie === 'function') { /* Reuse the early bootstrap cookie clearer when it is available. */
        SESSION_BOOTSTRAP.clearCookie(name); /* Delegate to the shared cookie clearer helper. */
        return; /* Stop after the shared clearer has expired the cookie. */
    }

    document.cookie = `${name}=; Max-Age=0; Path=/; SameSite=Lax`; /* Expire the cookie immediately across the whole site. */
}

function readTokens() { /* Read both auth tokens from localStorage. */
    const accessToken = window.localStorage.getItem(APP_STORAGE_KEYS.accessToken) || readCookie(APP_COOKIE_KEYS.accessToken) || ''; /* Prefer localStorage but fall back to the mirrored cookie when needed. */

    return { /* Return both token values as a small plain object. */
        accessToken: accessToken, /* Return the normalized access token. */
        refreshToken: '', /* Keep refresh-token reads empty because refresh now lives only in an HttpOnly cookie. */
    };
}

function writeTokens(tokens) { /* Persist a new token pair into localStorage. */
    const nextAccessToken = typeof tokens?.accessToken === 'string' ? tokens.accessToken.trim() : ''; /* Normalize the next access token before saving it. */

    if (nextAccessToken !== '') { /* Persist the access token only when it is non-empty. */
        window.localStorage.setItem(APP_STORAGE_KEYS.accessToken, nextAccessToken); /* Write the normalized access token to localStorage. */
        writeCookie(APP_COOKIE_KEYS.accessToken, nextAccessToken); /* Mirror the normalized access token into a site-wide cookie. */
        writeCookie(APP_COOKIE_KEYS.sessionMarker, '1'); /* Keep a lightweight readable session marker for SSR route gating. */
    } else { /* Handle an empty access token explicitly. */
        window.localStorage.removeItem(APP_STORAGE_KEYS.accessToken); /* Remove the stale access token from localStorage. */
        clearCookie(APP_COOKIE_KEYS.accessToken); /* Remove the stale access token cookie. */
        clearCookie(APP_COOKIE_KEYS.sessionMarker); /* Remove the readable session marker when access tokens are cleared. */
    }

}

function clearTokens() { /* Remove both auth tokens from localStorage. */
    window.localStorage.removeItem(APP_STORAGE_KEYS.accessToken); /* Delete the stored access token. */
    clearCookie(APP_COOKIE_KEYS.accessToken); /* Delete the mirrored access token cookie. */
    clearCookie(APP_COOKIE_KEYS.sessionMarker); /* Delete the readable session marker cookie. */
}

function hasSessionMarker() { /* Check whether the browser still holds the readable session marker cookie. */
    return readCookie(APP_COOKIE_KEYS.sessionMarker) !== ''; /* Treat any non-empty marker cookie as an active SSR session marker. */
}

function setStatus(node, message, state) { /* Update a live status node in a consistent way. */
    if (!node) { /* Guard against missing target nodes. */
        return; /* Stop early when the requested node does not exist on the current page. */
    }

    node.textContent = message || ''; /* Replace the visible message text. */

    if (state) { /* Apply a visual status marker when a state is provided. */
        node.dataset.state = state; /* Store the semantic state for CSS selectors. */
    } else { /* Handle an intentionally neutral state. */
        delete node.dataset.state; /* Remove any stale status state from the node. */
    }
}

function escapeHTML(value) { /* Escape raw text before injecting it into innerHTML. */
    return String(value) /* Normalize the incoming value into a string first. */
        .replaceAll('&', '&amp;') /* Escape ampersands before all other replacements. */
        .replaceAll('<', '&lt;') /* Escape opening angle brackets. */
        .replaceAll('>', '&gt;') /* Escape closing angle brackets. */
        .replaceAll('"', '&quot;') /* Escape double quotes. */
        .replaceAll("'", '&#39;'); /* Escape single quotes. */
}

async function readResponsePayload(response) { /* Safely parse the response body regardless of content type. */
    const contentType = response.headers.get('content-type') || ''; /* Read the response content type defensively. */

    if (contentType.includes('application/json')) { /* Prefer JSON parsing when the response advertises JSON. */
        return response.json().catch(() => null); /* Parse the JSON body or fall back to null on malformed JSON. */
    }

    return response.text().catch(() => ''); /* Fall back to a raw text body for non-JSON responses. */
}

async function apiRequest(path, options = {}) { /* Execute a fetch call against the backend API. */
    const method = options.method || 'GET'; /* Default every request to GET unless overridden. */
    const headers = new Headers(options.headers || {}); /* Normalize custom headers through the Headers API. */
    const requestInit = {method: method, headers: headers}; /* Build the base fetch configuration object. */
    const targetURL = `${API_BASE}${path}`; /* Combine the shared API prefix with the route fragment. */

    if (options.accessToken) { /* Attach the Authorization header when an access token exists. */
        headers.set('Authorization', `Bearer ${options.accessToken}`); /* Send the bearer token expected by auth-protected routes. */
    }

    if (options.body instanceof FormData) { /* Handle multipart requests explicitly. */
        requestInit.body = options.body; /* Pass the FormData instance through unchanged. */
    } else if (options.body !== undefined) { /* Handle JSON requests with a defined body. */
        headers.set('Content-Type', 'application/json'); /* Mark the request body as JSON. */
        requestInit.body = JSON.stringify(options.body); /* Serialize the request body into JSON text. */
    }

    const response = await fetch(targetURL, requestInit); /* Execute the HTTP request. */
    const payload = await readResponsePayload(response); /* Parse the response body using the shared parser. */

    return {response: response, payload: payload}; /* Return both the raw response and parsed payload to the caller. */
}

window.goHome = goHome; /* Expose the home helper for inline template buttons. */
window.goHomeToProfile = goHomeToProfile; /* Expose the profile-return helper for inline template buttons. */
window.goTo = goTo; /* Expose generic navigation for simple template links. */

window.AppWeb = { /* Expose a small shared runtime for page-specific scripts. */
    getCurrentPageName: getCurrentPageName, /* Share the current page name helper. */
    readTokens: readTokens, /* Share token reading logic. */
    hasSessionMarker: hasSessionMarker, /* Share the readable SSR session-marker helper. */
    writeTokens: writeTokens, /* Share token writing logic. */
    clearTokens: clearTokens, /* Share token clearing logic. */
    setStatus: setStatus, /* Share the live-status helper. */
    apiRequest: apiRequest, /* Share the fetch wrapper used by all pages. */
    escapeHTML: escapeHTML, /* Share the HTML escaping helper for safe innerHTML rendering. */
};

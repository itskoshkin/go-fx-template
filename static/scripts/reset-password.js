function resetT(key, fallback) { /* Resolve translated reset-page copy with a safe fallback when i18n is unavailable. */
    return window.AppI18n?.t(key, fallback) || fallback || key; /* Prefer runtime translations and fall back gracefully. */
}

document.addEventListener('DOMContentLoaded', () => { /* Wait for the DOM before querying reset-password nodes. */
    if (window.AppWeb.getCurrentPageName() !== 'reset-password') { /* Run this script only on the reset-password route. */
        return; /* Stop early on all other pages. */
    }

    const form = document.getElementById('reset-password-form'); /* Grab the reset-password form node. */
    const statusNode = document.getElementById('reset-password-status'); /* Grab the live status node for the form. */

    if (!form) { /* Guard against invalid reset-link pages that do not render the form. */
        return; /* Stop when there is no reset form to wire. */
    }

    form.addEventListener('submit', async (event) => { /* Handle password reset submission through fetch. */
        event.preventDefault(); /* Keep the browser from performing a native submit. */

        const token = form.elements.namedItem('token').value; /* Read the hidden reset token field. */
        const newPassword = form.elements.namedItem('new_password').value; /* Read the new password field. */
        const repeatPassword = form.elements.namedItem('repeat_password').value; /* Read the repeated password field. */

        if (newPassword !== repeatPassword) { /* Validate that both password fields match. */
            window.AppWeb.setStatus(statusNode, resetT('reset.passwordsMismatch', 'Passwords do not match.'), 'error'); /* Surface the mismatch immediately. */
            return; /* Stop before calling the API. */
        }

        window.AppWeb.setStatus(statusNode, resetT('reset.saving', 'Saving new password...'), 'pending'); /* Surface a pending state to the user. */

        try { /* Execute the password reset request and handle the result. */
            const result = await window.AppWeb.apiRequest('/auth/set-new-password', { /* Call the set-new-password endpoint. */
                method: 'POST', /* Send the request as POST. */
                body: {token: token, new_password: newPassword}, /* Serialize the token and new password into the JSON body. */
            });

            if (!result.response.ok) { /* Handle failed password reset attempts. */
                const message = result.payload?.message || resetT('reset.failed', 'Failed to reset password.'); /* Prefer the API error message when available. */
                window.AppWeb.setStatus(statusNode, message, 'error'); /* Surface the failure message. */
                return; /* Stop before resetting the form. */
            }

            const successMessage = result.payload?.message || resetT('reset.success', 'Password updated. You can return to the app and sign in.'); /* Prefer the API success message when available. */
            window.AppWeb.setStatus(statusNode, successMessage, 'success'); /* Surface the successful reset message. */
            form.reset(); /* Clear the visible password fields. */
            form.elements.namedItem('token').value = token; /* Restore the hidden token after reset clears the whole form. */
        } catch (_error) { /* Handle network-level failures. */
            window.AppWeb.setStatus(statusNode, resetT('auth.networkError', 'Network error. Please try again.'), 'error'); /* Surface a generic transport failure. */
        }
    });
});

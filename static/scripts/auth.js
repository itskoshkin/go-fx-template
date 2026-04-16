document.addEventListener('DOMContentLoaded', () => { /* Wait for the DOM before querying auth-page nodes. */
    const pageName = window.AppWeb.getCurrentPageName(); /* Read the logical page marker from the body. */

    if (pageName === 'login') { /* Initialize the unified auth page only on the login route. */
        initAuthPage(); /* Wire the login/register page behavior. */
    }
});

function authT(key, fallback) { /* Resolve translated UI copy with a safe fallback for pages that load before i18n is ready. */
    return window.AppI18n?.t(key, fallback) || fallback || key; /* Prefer runtime translations and fall back gracefully. */
}

function initAuthPage() { /* Attach the unified login/register page behavior. */
    const tokens = window.AppWeb.readTokens(); /* Read the currently stored token pair. */
    const hasSessionMarker = window.AppWeb.hasSessionMarker(); /* Check whether the browser still has a readable session marker cookie. */

    if (tokens.accessToken !== '' || hasSessionMarker) { /* Redirect away from auth when a session already exists locally or via cookie-backed refresh state. */
        window.goHome(); /* Send the user to the authenticated dashboard route. */
        return; /* Stop before wiring the auth page. */
    }

    const root = document.body; /* Use the body as the source of initial auth-page metadata. */
    const elements = { /* Cache auth-page DOM nodes used throughout the auth flow. */
        showLoginViewButton: document.getElementById('show-login-view-btn'), /* Cache the login segment button. */
        showRegisterViewButton: document.getElementById('show-register-view-btn'), /* Cache the register segment button. */
        loginForm: document.getElementById('login-form'), /* Cache the login form. */
        registerForm: document.getElementById('register-form'), /* Cache the register form. */
        loginFormTitle: document.getElementById('login-form-title'), /* Cache the login form heading node. */
        loginCredentialsFields: document.getElementById('login-credentials-fields'), /* Cache the username/password field group. */
        forgotPasswordFields: document.getElementById('forgot-password-fields'), /* Cache the forgot-password email field group. */
        loginUsernameInput: document.getElementById('login-username'), /* Cache the login username input. */
        loginPasswordInput: document.getElementById('login-password'), /* Cache the login password input. */
        forgotPasswordEmailInput: document.getElementById('forgot-password-email'), /* Cache the forgot-password email input. */
        loginSubmitButton: document.getElementById('login-submit-btn'), /* Cache the primary login-form submit button. */
        toggleForgotPasswordButton: document.getElementById('toggle-forgot-password-btn'), /* Cache the forgot-password toggle button. */
        backToLoginButton: document.getElementById('back-to-login-btn'), /* Cache the back-to-login button. */
        loginStatus: document.getElementById('login-status'), /* Cache the login-form status line. */
        registerStatus: document.getElementById('register-status'), /* Cache the register-form status line. */
    };

    const state = { /* Track the current auth-page UI state. */
        authView: 'login', /* Track whether login or register is visible. */
        forgotPasswordMode: false, /* Track whether the login form is acting as forgot-password form. */
    };

    function setAuthView(nextView) { /* Toggle between login and register forms. */
        state.authView = nextView; /* Store the currently visible auth view. */
        state.forgotPasswordMode = false; /* Reset forgot-password submode when switching top-level view. */
        elements.loginForm.hidden = nextView !== 'login'; /* Show the login form only in login view. */
        elements.registerForm.hidden = nextView !== 'register'; /* Show the register form only in register view. */
        elements.showLoginViewButton.classList.toggle('is-active', nextView === 'login'); /* Highlight the login segment when active. */
        elements.showRegisterViewButton.classList.toggle('is-active', nextView === 'register'); /* Highlight the register segment when active. */
        syncForgotPasswordMode(); /* Re-render login-form submode after the top-level view changes. */
        window.AppWeb.setStatus(elements.loginStatus, '', undefined); /* Clear stale login-form status text. */
        window.AppWeb.setStatus(elements.registerStatus, '', undefined); /* Clear stale register-form status text. */
    }

    function syncForgotPasswordMode() { /* Re-render the login form based on current forgot-password state. */
        const enabled = state.forgotPasswordMode; /* Read the current forgot-password submode flag. */
        elements.loginFormTitle.textContent = enabled ? authT('auth.formTitleForgotPassword', 'Forgot Password') : authT('auth.formTitleLogin', 'Log In'); /* Update the login form heading for the current submode. */
        elements.loginCredentialsFields.hidden = enabled; /* Hide username/password fields in forgot-password mode. */
        elements.forgotPasswordFields.hidden = !enabled; /* Show the forgot-password email field only in forgot mode. */
        elements.toggleForgotPasswordButton.hidden = enabled; /* Hide the forgot-password trigger while forgot mode is active. */
        elements.backToLoginButton.hidden = !enabled; /* Show the back button only in forgot mode. */
        elements.loginSubmitButton.textContent = enabled ? authT('auth.sendResetLink', 'Send Reset Link') : authT('auth.login', 'Log In'); /* Update the submit label for the current submode. */
        elements.loginUsernameInput.required = !enabled; /* Require username only in normal login mode. */
        elements.loginPasswordInput.required = !enabled; /* Require password only in normal login mode. */
        elements.forgotPasswordEmailInput.required = enabled; /* Require email only in forgot-password mode. */
    }

    elements.showLoginViewButton.addEventListener('click', () => { /* Switch to log in mode on segmented-control click. */
        setAuthView('login'); /* Show the login form. */
    });

    elements.showRegisterViewButton.addEventListener('click', () => { /* Switch to register mode on segmented-control click. */
        setAuthView('register'); /* Show the register form. */
    });

    elements.toggleForgotPasswordButton.addEventListener('click', () => { /* Enter forgot-password mode from the login form. */
        state.forgotPasswordMode = true; /* Enable the forgot-password submode. */
        syncForgotPasswordMode(); /* Re-render the login form for forgot-password mode. */
        window.AppWeb.setStatus(elements.loginStatus, '', undefined); /* Clear stale login status when switching submode. */
    });

    elements.backToLoginButton.addEventListener('click', () => { /* Return from forgot-password mode to normal login mode. */
        state.forgotPasswordMode = false; /* Disable forgot-password submode. */
        syncForgotPasswordMode(); /* Re-render the login form for normal login mode. */
        window.AppWeb.setStatus(elements.loginStatus, '', undefined); /* Clear stale forgot-password status when returning. */
    });

    elements.loginForm.addEventListener('submit', async (event) => { /* Handle login and forgot-password submissions through one form. */
        event.preventDefault(); /* Keep the browser from performing a native form submit. */

        if (state.forgotPasswordMode) { /* Handle forgot-password requests inside the login form. */
            const email = elements.forgotPasswordEmailInput.value.trim(); /* Read the forgot-password email field. */
            window.AppWeb.setStatus(elements.loginStatus, authT('auth.submittingResetRequest', 'Submitting reset request...'), 'pending'); /* Surface a pending forgot-password state. */

            try { /* Execute the forgot-password request and handle the result. */
                const result = await window.AppWeb.apiRequest('/auth/forgot-password', { /* Call the forgot-password endpoint. */
                    method: 'POST', /* Send the request as POST. */
                    body: {email: email}, /* Serialize the email payload into JSON. */
                });

                if (!result.response.ok) { /* Handle failed forgot-password attempts. */
                    const message = result.payload?.message || authT('auth.resetRequestFailed', 'Password reset request failed.'); /* Prefer the API error message when available. */
                    window.AppWeb.setStatus(elements.loginStatus, message, 'error'); /* Surface the forgot-password failure. */
                    return; /* Stop before clearing fields. */
                }

                elements.forgotPasswordEmailInput.value = ''; /* Clear the forgot-password email field after success. */
                window.AppWeb.setStatus(elements.loginStatus, result.payload?.message || authT('auth.resetRequestSuccess', 'If the account exists, an email will arrive shortly.'), 'success'); /* Surface the neutral forgot-password success response. */
            } catch (_error) { /* Handle network-level failures. */
                window.AppWeb.setStatus(elements.loginStatus, authT('auth.networkError', 'Network error. Please try again.'), 'error'); /* Surface a generic transport failure. */
            }

            return; /* Stop before evaluating normal login flow. */
        }

        const username = elements.loginUsernameInput.value.trim(); /* Read the login username field. */
        const password = elements.loginPasswordInput.value; /* Read the login password field. */
        window.AppWeb.setStatus(elements.loginStatus, authT('auth.loggingIn', 'Logging in...'), 'pending'); /* Surface a pending login state. */

        try { /* Execute the login request and handle the result. */
            const result = await window.AppWeb.apiRequest('/auth/login', { /* Call the login endpoint. */
                method: 'POST', /* Send the request as POST. */
                body: {username: username, password: password}, /* Serialize the login payload into JSON. */
            });

            if (!result.response.ok) { /* Handle failed login attempts. */
                const message = result.payload?.message || authT('auth.loginFailed', 'Login failed.'); /* Prefer the API error message when available. */
                window.AppWeb.setStatus(elements.loginStatus, message, 'error'); /* Surface the login failure. */
                return; /* Stop before writing local session state. */
            }

            window.AppWeb.writeTokens({ /* Persist the token pair returned by the backend. */
                accessToken: result.payload?.['access_token'] || '', /* Save the returned access token. */
            });

            window.goHome(); /* Redirect to the authenticated dashboard after login. */
        } catch (_error) { /* Handle network-level failures. */
            window.AppWeb.setStatus(elements.loginStatus, authT('auth.networkError', 'Network error. Please try again.'), 'error'); /* Surface a generic transport failure. */
        }
    });

    elements.registerForm.addEventListener('submit', async (event) => { /* Handle registration through the register form. */
        event.preventDefault(); /* Keep the browser from performing a native form submit. */

        const payload = { /* Build the register payload from current form values. */
            name: elements.registerForm.elements.namedItem('name').value.trim(), /* Read the submitted display name. */
            username: elements.registerForm.elements.namedItem('username').value.trim(), /* Read the submitted username. */
            password: elements.registerForm.elements.namedItem('password').value, /* Read the submitted password. */
        };

        const email = elements.registerForm.elements.namedItem('email').value.trim(); /* Read the optional email field. */

        if (email !== '') { /* Include email only when the field is non-empty. */
            payload.email = email; /* Add the email to the register payload. */
        }

        window.AppWeb.setStatus(elements.registerStatus, authT('auth.creatingAccount', 'Creating account...'), 'pending'); /* Surface a pending register state. */

        try { /* Execute the register request and handle the result. */
            const result = await window.AppWeb.apiRequest('/auth/register', { /* Call the register endpoint. */
                method: 'POST', /* Send the request as POST. */
                body: payload, /* Serialize the register payload into JSON. */
            });

            if (!result.response.ok) { /* Handle failed registration attempts. */
                const message = result.payload?.message || authT('auth.registrationFailed', 'Registration failed.'); /* Prefer the API error message when available. */
                window.AppWeb.setStatus(elements.registerStatus, message, 'error'); /* Surface the register failure. */
                return; /* Stop before writing local session state. */
            }

            window.AppWeb.writeTokens({ /* Persist the token pair returned by the backend. */
                accessToken: result.payload?.['access_token'] || '', /* Save the returned access token. */
            });

            window.goHome(); /* Redirect to the authenticated dashboard after registration. */
        } catch (_error) { /* Handle network-level failures. */
            window.AppWeb.setStatus(elements.registerStatus, authT('auth.networkError', 'Network error. Please try again.'), 'error'); /* Surface a generic transport failure. */
        }
    });

    setAuthView(root.dataset.initialAuthView === 'register' ? 'register' : 'login'); /* Apply the initial auth view provided by the server route. */

    if (root.dataset.initialForgotPassword === 'true' && state.authView === 'login') { /* Enter forgot-password mode when requested by the route. */
        state.forgotPasswordMode = true; /* Enable forgot-password submode. */
        syncForgotPasswordMode(); /* Re-render the login form for forgot-password mode. */
    }

    document.addEventListener('app:languagechange', syncForgotPasswordMode); /* Keep the auth form title and submit label aligned with live language switches. */
}

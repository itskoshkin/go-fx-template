document.addEventListener('DOMContentLoaded', () => { /* Wait for the DOM before wiring dashboard nodes. */
    const pageName = window.AppWeb.getCurrentPageName(); /* Read the logical page marker from the current body node. */

    if (pageName === 'dashboard') { /* Initialize the item dashboard only on the dashboard route. */
        initDashboardPage(); /* Attach the authenticated dashboard behavior. */
        return; /* Stop before evaluating other page initializers. */
    }

    if (pageName === 'change-password') { /* Initialize the standalone change-password page only on that route. */
        initChangePasswordPage(); /* Attach the authenticated current/new password form behavior. */
        return; /* Stop before evaluating other page initializers. */
    }

    if (pageName === 'delete-account-page') { /* Initialize the standalone delete-account page only on that route. */
        initDeleteAccountPage(); /* Attach the authenticated delete-account form behavior. */
    }
});

function dashboardT(key, fallback) { /* Resolve translated dashboard copy with a safe fallback when i18n is unavailable. */
    return window.AppI18n?.t(key, fallback) || fallback || key; /* Prefer runtime translations and fall back gracefully. */
}

function dashboardFormat(key, params, fallback) { /* Resolve translated dashboard copy with placeholder interpolation. */
    return window.AppI18n?.format(key, params, fallback) || fallback || key; /* Prefer runtime translation formatting and fall back gracefully. */
}

function initDashboardPage() { /* Wire the authenticated items dashboard. */
    const state = { /* Keep mutable dashboard state in one place. */
        currentUser: null, /* Store the currently authenticated user payload. */
        listedItems: [], /* Store the latest paginated item list. */
        pagination: {page: 1, limit: 10, totalObjects: 0, totalPages: 1}, /* Track current paging state. */
        itemModalMode: 'create', /* Track whether the item modal is creating or editing. */
        itemPersistedImageURL: '', /* Store the persisted item-image URL currently shown in the modal. */
        itemPreviewObjectURL: '', /* Store a temporary blob URL for the currently selected unsaved item image. */
        pendingDeleteItemID: '', /* Store the item ID currently awaiting delete confirmation. */
        closeItemModalAfterDelete: false, /* Track whether deleting from the confirm modal should also close the edit modal. */
    };

    const elements = { /* Cache dashboard DOM nodes that are reused throughout the page. */
        openProfileModalButton: document.getElementById('open-profile-modal-btn'), /* Cache the profile icon button. */
        logoutButton: document.getElementById('logout-btn'), /* Cache the logout icon button. */
        openCreateItemModalButton: document.getElementById('open-create-item-modal-btn'), /* Cache the new-item button. */
        dashboardStatus: document.getElementById('dashboard-status'), /* Cache the global dashboard status line. */
        itemsControlsForm: document.getElementById('items-controls-form'), /* Cache the paging form. */
        prevPageButton: document.getElementById('prev-page-btn'), /* Cache the previous-page button. */
        nextPageButton: document.getElementById('next-page-btn'), /* Cache the next-page button. */
        itemsPageInput: document.getElementById('items-page'), /* Cache the current page input. */
        itemsLimitInput: document.getElementById('items-limit'), /* Cache the page-size input. */
        itemsSummary: document.getElementById('items-summary'), /* Cache the items summary line. */
        itemsList: document.getElementById('items-list'), /* Cache the rendered item-card container. */
        itemsPaginationPages: document.getElementById('items-pagination-pages'), /* Cache the numbered pagination button container. */
        dashboardToast: document.getElementById('dashboard-toast'), /* Cache the bottom toast node. */
        itemModal: document.getElementById('item-modal'), /* Cache the item dialog shell. */
        closeItemModalButton: document.getElementById('close-item-modal-btn'), /* Cache the item-modal close button. */
        itemModalModeLabel: document.getElementById('item-modal-mode-label'), /* Cache the item-modal eyebrow label. */
        itemModalTitle: document.getElementById('item-modal-title'), /* Cache the item-modal title. */
        itemForm: document.getElementById('item-form'), /* Cache the item create/edit form. */
        itemIDInput: document.getElementById('item-id'), /* Cache the hidden item ID input. */
        itemTitleInput: document.getElementById('item-title'), /* Cache the item title input. */
        itemDescriptionInput: document.getElementById('item-description'), /* Cache the item description textarea. */
        itemImagePickerButton: document.getElementById('item-image-picker-btn'), /* Cache the clickable item-image circle button. */
        itemImagePreview: document.getElementById('item-image-preview'), /* Cache the item-image preview node inside the circle. */
        itemImageFileInput: document.getElementById('item-image-file'), /* Cache the item image file input. */
        removeItemImageButton: document.getElementById('remove-item-image-btn'), /* Cache the remove-item-image button. */
        deleteItemButton: document.getElementById('delete-item-btn'), /* Cache the delete-item button. */
        itemModalStatus: document.getElementById('item-modal-status'), /* Cache the item-modal status line. */
        deleteItemConfirmModal: document.getElementById('delete-item-confirm-modal'), /* Cache the delete-item confirm dialog shell. */
        closeDeleteItemConfirmModalButton: document.getElementById('close-delete-item-confirm-modal-btn'), /* Cache the delete-item confirm close button. */
        cancelDeleteItemButton: document.getElementById('cancel-delete-item-btn'), /* Cache the delete-item confirm cancel button. */
        confirmDeleteItemButton: document.getElementById('confirm-delete-item-btn'), /* Cache the delete-item confirm destructive button. */
        profileModal: document.getElementById('profile-modal'), /* Cache the profile dialog shell. */
        closeProfileModalButton: document.getElementById('close-profile-modal-btn'), /* Cache the profile-modal close button. */
        profileForm: document.getElementById('profile-form'), /* Cache the profile form. */
        avatarPickerButton: document.getElementById('avatar-picker-btn'), /* Cache the clickable avatar circle button. */
        avatarPlaceholder: document.getElementById('avatar-placeholder'), /* Cache the avatar placeholder artwork shown when no avatar exists. */
        avatarPreview: document.getElementById('avatar-preview'), /* Cache the avatar preview image inside the circle. */
        avatarFileInput: document.getElementById('avatar-file'), /* Cache the hidden avatar file input. */
        deleteAvatarButton: document.getElementById('delete-avatar-btn'), /* Cache the delete-avatar button next to the circle. */
        profileNameInput: document.getElementById('profile-name'), /* Cache the profile name input. */
        profileUsernameInput: document.getElementById('profile-username'), /* Cache the profile username input. */
        profileEmailInput: document.getElementById('profile-email'), /* Cache the profile email input. */
        profileStatus: document.getElementById('profile-status'), /* Cache the profile status line. */
        openChangePasswordPageButton: document.getElementById('open-change-password-page-btn'), /* Cache the change-password navigation button. */
        openDeleteAccountPageButton: document.getElementById('open-delete-account-page-btn'), /* Cache the delete-account navigation button. */
    };

    syncPaginationStateFromInputs(elements, state); /* Normalize the initial paging inputs into state. */
    bindDashboardEvents(elements, state); /* Wire all dashboard events. */
    document.addEventListener('app:languagechange', () => { /* Re-render dynamic dashboard copy when the active UI language changes. */
        rerenderDashboardLocale(elements, state); /* Refresh summaries, cards, modal labels, and tooltips in place. */
    });
    void bootstrapDashboard(elements, state); /* Restore the current session or redirect to login. */
}

function rerenderDashboardLocale(elements, state) { /* Refresh dynamic dashboard copy that is rendered from JavaScript instead of templates. */
    updatePagingSummary(elements, state); /* Refresh the compact pagination summary line in the current language. */
    renderItemsList(elements, state); /* Re-render item cards so action labels and date captions change language too. */

    if (state.itemModalMode === 'edit') { /* Keep the open item modal eyebrow aligned with the current mode and language. */
        elements.itemModalModeLabel.textContent = dashboardT('item.edit', 'Edit'); /* Translate the edit-mode eyebrow label in place. */
    } else { /* Treat every non-edit state as create mode. */
        elements.itemModalModeLabel.textContent = dashboardT('item.create', 'Create'); /* Translate the create-mode eyebrow label in place. */
    }

    if (state.currentUser) { /* Refresh avatar tooltip copy when user data is already loaded. */
        syncProfileAvatarState(elements, state.currentUser); /* Re-render the profile avatar state so its tooltip matches the active language. */
    }

    renderItemImageState(elements, state, state.itemPreviewObjectURL || state.itemPersistedImageURL); /* Refresh the item-image circle tooltip in the active language. */
}

function bindDashboardEvents(elements, state) { /* Attach all dashboard interactions. */
    elements.openProfileModalButton.addEventListener('click', async () => { /* Open the profile modal on demand. */
        await openProfileModal(elements, state); /* Load the latest current user and open the profile dialog. */
    });

    elements.logoutButton.addEventListener('click', async () => { /* Log out the current session on demand. */
        await logOutAndRedirect(elements); /* Execute logout flow and return to the login route. */
    });

    elements.openCreateItemModalButton.addEventListener('click', () => { /* Open the create-item modal on demand. */
        openCreateItemModal(elements, state); /* Reset the item modal into create mode. */
    });

    elements.itemsControlsForm.addEventListener('submit', async (event) => { /* Apply paging values through the paging form. */
        event.preventDefault(); /* Keep the browser from performing a native form submit. */
        syncPaginationStateFromInputs(elements, state); /* Normalize the current paging inputs into state. */
        await loadItems(elements, state); /* Reload items using the new paging values. */
    });

    elements.prevPageButton.addEventListener('click', async () => { /* Load the previous page on demand. */
        if (state.pagination.page <= 1) { /* Guard against invalid backward paging. */
            return; /* Stop before mutating the current page. */
        }

        state.pagination.page -= 1; /* Decrement the current page number. */
        elements.itemsPageInput.value = String(state.pagination.page); /* Mirror the new page back into the input. */
        await loadItems(elements, state); /* Reload items for the previous page. */
    });

    elements.nextPageButton.addEventListener('click', async () => { /* Load the next page on demand. */
        if (state.pagination.page >= state.pagination.totalPages) { /* Guard against invalid forward paging. */
            return; /* Stop before mutating the current page. */
        }

        state.pagination.page += 1; /* Increment the current page number. */
        elements.itemsPageInput.value = String(state.pagination.page); /* Mirror the new page back into the input. */
        await loadItems(elements, state); /* Reload items for the next page. */
    });

    elements.itemsPaginationPages.addEventListener('click', async (event) => { /* Jump directly to a numbered page on demand. */
        const button = event.target.closest('[data-page-number]'); /* Detect a numbered pagination button in the click path. */

        if (!button) { /* Ignore clicks outside numbered pagination controls. */
            return; /* Stop before reading pagination data. */
        }

        const nextPage = Number(button.dataset.pageNumber) || 1; /* Read the requested page number from the clicked button. */

        if (nextPage === state.pagination.page) { /* Ignore clicks on the already active page. */
            return; /* Stop before reloading the same page. */
        }

        state.pagination.page = nextPage; /* Store the requested page number in dashboard state. */
        elements.itemsPageInput.value = String(state.pagination.page); /* Mirror the requested page into the top control field. */
        await loadItems(elements, state); /* Reload items for the selected numbered page. */
    });

    elements.itemsList.addEventListener('click', async (event) => { /* Use event delegation for item-card actions. */
        const button = event.target.closest('[data-action]'); /* Look for an action button in the click path. */

        if (!button) { /* Ignore clicks unrelated to item actions. */
            return; /* Stop when the click target has no dashboard action. */
        }

        const itemID = button.dataset.itemId || ''; /* Read the target item ID from the clicked button. */

        if (itemID === '') { /* Guard against malformed action buttons. */
            return; /* Stop when the clicked action carries no item ID. */
        }

        if (button.dataset.action === 'edit-item') { /* Handle edit-item actions. */
            await openEditItemModal(elements, state, itemID); /* Load the item and open the edit modal. */
            return; /* Stop before evaluating other action branches. */
        }

        if (button.dataset.action === 'delete-item') { /* Handle delete-item actions. */
            openDeleteItemConfirmModal(elements, state, itemID, false); /* Open the custom delete-confirm modal for list-row deletes. */
        }
    });

    elements.closeItemModalButton.addEventListener('click', () => { /* Close the item modal on demand. */
        closeDialog(elements.itemModal); /* Close the item dialog shell safely. */
    });

    elements.itemForm.addEventListener('submit', async (event) => { /* Save the current item modal contents on submit. */
        event.preventDefault(); /* Keep the browser from performing a native form submit. */
        await saveItemFromModal(elements, state); /* Create or update the current item based on modal mode. */
    });

    elements.itemImagePickerButton.addEventListener('click', () => { /* Open the hidden item-image file picker when the image circle is clicked. */
        elements.itemImageFileInput.click(); /* Forward item-image circle clicks to the hidden file input. */
    });

    elements.itemImageFileInput.addEventListener('change', () => { /* Preview a newly selected item image immediately inside the image circle. */
        syncItemImagePreviewFromSelection(elements, state); /* Mirror the current file-input selection into the item-image circle. */
    });

    elements.removeItemImageButton.addEventListener('click', async () => { /* Remove the current item's image on demand. */
        if (elements.itemImageFileInput.files[0]) { /* Clear pending unsaved image selections locally before calling backend endpoints. */
            elements.itemImageFileInput.value = ''; /* Clear the pending file-input selection. */
            syncItemImageFromPersistedURL(elements, state, state.itemPersistedImageURL); /* Restore the persisted image state or the empty circle. */
            return; /* Stop before calling the remove-image endpoint. */
        }

        const itemID = elements.itemIDInput.value.trim(); /* Read the current modal item ID. */

        if (itemID === '' || state.itemPersistedImageURL === '') { /* Guard against removing an image in create mode or when no persisted image exists. */
            return; /* Stop when there is no persisted item yet. */
        }

        await removeItemImage(elements, state, itemID); /* Remove the item image and refresh the modal payload. */
    });

    elements.deleteItemButton.addEventListener('click', async () => { /* Delete the current modal item on demand. */
        const itemID = elements.itemIDInput.value.trim(); /* Read the current modal item ID. */

        if (itemID === '') { /* Guard against invalid deletes without an item ID. */
            return; /* Stop when the item cannot or should not be deleted. */
        }

        openDeleteItemConfirmModal(elements, state, itemID, true); /* Open the custom delete-confirm modal for modal-based deletes. */
    });

    elements.itemModal.addEventListener('close', () => { /* Reset transient item-modal state after the dialog closes. */
        window.AppWeb.setStatus(elements.itemModalStatus, '', undefined); /* Clear the item-modal status line. */
        elements.itemImageFileInput.value = ''; /* Clear the pending item image file field. */
        syncItemImageFromPersistedURL(elements, state, ''); /* Reset the item-image circle after the modal fully closes. */
    });

    elements.itemModal.addEventListener('click', (event) => { /* Support backdrop-style closing for fallback dialogs. */
        if (event.target === elements.itemModal) { /* Detect clicks on the dialog shell instead of the inner surface. */
            closeDialog(elements.itemModal); /* Close the item modal when its backdrop area is clicked. */
        }
    });

    elements.closeDeleteItemConfirmModalButton.addEventListener('click', () => { /* Close the delete-item confirm modal on demand. */
        closeDialog(elements.deleteItemConfirmModal); /* Close the custom delete-item confirm dialog safely. */
    });

    elements.cancelDeleteItemButton.addEventListener('click', () => { /* Cancel item deletion from the custom confirm modal. */
        closeDialog(elements.deleteItemConfirmModal); /* Close the custom delete-item confirm dialog safely. */
    });

    elements.confirmDeleteItemButton.addEventListener('click', async () => { /* Confirm item deletion from the custom confirm modal. */
        const itemID = state.pendingDeleteItemID; /* Read the pending item ID currently staged for deletion. */

        if (itemID === '') { /* Guard against stale confirm clicks without a staged item ID. */
            closeDialog(elements.deleteItemConfirmModal); /* Close the stale confirm dialog instead of doing nothing visibly. */
            return; /* Stop before calling delete endpoints. */
        }

        closeDialog(elements.deleteItemConfirmModal); /* Close the confirm dialog before executing the delete flow. */
        const deleted = await deleteItemByID(elements, state, itemID); /* Delete the staged item and refresh the list. */

        if (deleted && state.closeItemModalAfterDelete) { /* Close the edit modal too when the delete started from inside it. */
            closeDialog(elements.itemModal); /* Close the item edit modal after the delete completes successfully. */
        }
    });

    elements.deleteItemConfirmModal.addEventListener('close', () => { /* Reset transient delete-confirm modal state after it closes. */
        resetDeleteItemConfirmState(state); /* Clear the staged delete target after the confirm modal closes. */
    });

    elements.deleteItemConfirmModal.addEventListener('click', (event) => { /* Support backdrop-style closing for the delete-confirm modal. */
        if (event.target === elements.deleteItemConfirmModal) { /* Detect clicks on the confirm dialog shell instead of the inner surface. */
            closeDialog(elements.deleteItemConfirmModal); /* Close the delete-confirm modal when its backdrop area is clicked. */
        }
    });

    elements.closeProfileModalButton.addEventListener('click', () => { /* Close the profile modal on demand. */
        closeDialog(elements.profileModal); /* Close the profile dialog shell safely. */
    });

    elements.profileModal.addEventListener('click', (event) => { /* Support backdrop-style closing for fallback dialogs. */
        if (event.target === elements.profileModal) { /* Detect clicks on the dialog shell instead of the inner surface. */
            closeDialog(elements.profileModal); /* Close the profile modal when its backdrop area is clicked. */
        }
    });

    elements.profileForm.addEventListener('submit', async (event) => { /* Save profile fields on submit. */
        event.preventDefault(); /* Keep the browser from performing a native form submit. */
        await saveProfile(elements, state); /* PATCH the current user profile. */
    });

    elements.avatarPickerButton.addEventListener('click', () => { /* Open the hidden avatar file picker when the avatar circle is clicked. */
        elements.avatarFileInput.click(); /* Forward avatar-circle clicks to the hidden file input. */
    });

    elements.avatarFileInput.addEventListener('change', async () => { /* Upload a new avatar immediately after file selection. */
        if (!elements.avatarFileInput.files[0]) { /* Ignore empty file-input changes. */
            return; /* Stop before attempting avatar upload. */
        }

        await uploadAvatar(elements, state); /* PUT the newly selected avatar image for the current user. */
    });

    elements.deleteAvatarButton.addEventListener('click', async () => { /* Remove the current avatar on demand. */
        await deleteAvatar(elements, state); /* DELETE the current user avatar. */
    });

    elements.openChangePasswordPageButton.addEventListener('click', () => { /* Navigate to the standalone change-password page on demand. */
        window.goTo('/change-password'); /* Open the current/new password page outside the modal. */
    });

    elements.openDeleteAccountPageButton.addEventListener('click', () => { /* Navigate to the standalone delete-account page on demand. */
        window.goTo('/delete-account'); /* Open the delete-account page outside the modal. */
    });

    elements.avatarPreview.addEventListener('error', () => { /* Recover cleanly when the current avatar URL fails to load. */
        elements.avatarPlaceholder.hidden = false; /* Restore the avatar placeholder artwork when the preview image fails to load. */
        elements.avatarPreview.hidden = true; /* Hide the broken avatar image element. */
        elements.avatarPreview.removeAttribute('src'); /* Remove the failed avatar source URL. */
        elements.avatarPickerButton.classList.remove('has-image'); /* Restore the empty-circle avatar state. */
        elements.deleteAvatarButton.hidden = true; /* Hide the delete-avatar control when the current image cannot be rendered. */
    });

    elements.itemImagePreview.addEventListener('error', () => { /* Recover cleanly when the current item image URL fails to load. */
        elements.itemImagePreview.hidden = true; /* Hide the broken item-image element. */
        elements.itemImagePreview.removeAttribute('src'); /* Remove the failed item-image source URL. */
        elements.itemImagePickerButton.classList.remove('has-image'); /* Restore the empty-circle item-image state. */
    });
}

function initChangePasswordPage() { /* Wire the standalone current/new password page. */
    const form = document.getElementById('change-password-page-form'); /* Grab the standalone change-password form node. */
    const statusNode = document.getElementById('change-password-page-status'); /* Grab the live status node for the page. */

    if (!form) { /* Guard against missing markup on the current route. */
        return; /* Stop when the change-password form is absent. */
    }

    form.addEventListener('submit', async (event) => { /* Change the current user's password from the standalone page. */
        event.preventDefault(); /* Keep the browser from performing a native form submit. */

        const currentPassword = form.elements.namedItem('current_password').value; /* Read the current password field. */
        const newPassword = form.elements.namedItem('new_password').value; /* Read the new password field. */

        if (currentPassword === '' || newPassword === '') { /* Validate that both password fields are present. */
            window.AppWeb.setStatus(statusNode, dashboardT('changePassword.fillBoth', 'Fill both password fields.'), 'error'); /* Surface the incomplete-password error immediately. */
            return; /* Stop before calling protected endpoints. */
        }

        window.AppWeb.setStatus(statusNode, dashboardT('changePassword.pending', 'Changing password...'), 'pending'); /* Surface a pending password-change state. */
        const result = await runAuthorizedRequest('/users/me/update-password', {method: 'PATCH', body: {current_password: currentPassword, new_password: newPassword}}, statusNode); /* Call the current-password update endpoint with session recovery. */

        if (!result.response) { /* Handle expired or missing sessions during password change. */
            redirectToLogin(true); /* Return to the login route after clearing stale tokens. */
            return; /* Stop before mutating the standalone page state. */
        }

        if (!result.response.ok) { /* Handle non-auth password-change failures. */
            window.AppWeb.setStatus(statusNode, result.payload?.message || dashboardT('changePassword.failed', 'Failed to change password.'), 'error'); /* Surface the password-change failure. */
            return; /* Stop before redirecting back to the dashboard. */
        }

        window.AppWeb.setStatus(statusNode, '', undefined); /* Keep the inline status line empty after a successful password change. */
        form.reset(); /* Clear both password fields after success. */

        window.setTimeout(() => { /* Give the user a moment to register the success message. */
            window.goHome(); /* Return to the dashboard after a successful password change. */
        }, 900); /* Keep the success message visible briefly before redirecting. */
    });
}

function initDeleteAccountPage() { /* Wire the standalone delete-account page. */
    const form = document.getElementById('delete-account-page-form'); /* Grab the standalone delete-account form node. */
    const statusNode = document.getElementById('delete-account-page-status'); /* Grab the live status node for the page. */

    if (!form) { /* Guard against missing markup on the current route. */
        return; /* Stop when the delete-account form is absent. */
    }

    form.addEventListener('submit', async (event) => { /* Delete the authenticated user from the standalone page. */
        event.preventDefault(); /* Keep the browser from performing a native form submit. */

        const currentPassword = form.elements.namedItem('password').value; /* Read the current-password confirmation field. */

        if (currentPassword === '') { /* Validate that the password confirmation field is not empty. */
            window.AppWeb.setStatus(statusNode, dashboardT('deleteAccount.passwordRequired', 'Current password is required.'), 'error'); /* Surface the missing-password error immediately. */
            return; /* Stop before calling protected endpoints. */
        }

        window.AppWeb.setStatus(statusNode, dashboardT('deleteAccount.pending', 'Deleting account...'), 'pending'); /* Surface a pending account-deletion state. */
        const result = await runAuthorizedRequest('/users/me', {method: 'DELETE', body: {password: currentPassword}}, statusNode); /* Call the delete-account endpoint with session recovery. */

        if (!result.response) { /* Handle expired or missing sessions during account deletion. */
            redirectToLogin(true); /* Return to the login route after clearing stale tokens. */
            return; /* Stop before mutating standalone page state. */
        }

        if (!result.response.ok && result.response.status !== 204) { /* Handle non-auth account-deletion failures. */
            window.AppWeb.setStatus(statusNode, result.payload?.message || dashboardT('deleteAccount.failed', 'Failed to delete account.'), 'error'); /* Surface the account-deletion failure. */
            return; /* Stop before clearing the session. */
        }

        window.AppWeb.clearTokens(); /* Clear both local auth tokens after the account has been deleted. */
        window.goTo('/login'); /* Return to the login route after account deletion completes. */
    });
}

async function bootstrapDashboard(elements, state) { /* Restore the current session or redirect to login. */
    const tokens = window.AppWeb.readTokens(); /* Read the currently stored token pair. */
    const hasSessionMarker = window.AppWeb.hasSessionMarker(); /* Check whether the browser still has a readable session marker cookie. */

    if (tokens.accessToken === '' && !hasSessionMarker) { /* Handle a fully missing local session when neither access nor cookie-backed refresh state exists. */
        redirectToLogin(true); /* Return to the login route immediately. */
        return; /* Stop before calling protected routes. */
    }

    const userLoaded = await loadCurrentUser(elements, state); /* Load the current user before rendering the dashboard. */

    if (!userLoaded) { /* Handle sessions that cannot load the current user. */
        return; /* Stop before requesting the item list. */
    }

    await loadItems(elements, state); /* Load the current page of items for the authenticated user. */
    await maybeOpenProfileModalFromURL(elements, state); /* Restore the profile modal when the route explicitly asks for it. */
}

async function maybeOpenProfileModalFromURL(elements, state) { /* Reopen the profile modal automatically when the dashboard route requests it. */
    const currentURL = new URL(window.location.href); /* Read the current browser URL so query parameters can drive modal state. */

    if (currentURL.searchParams.get('modal') !== 'profile') { /* Ignore dashboard loads that do not explicitly target the profile modal. */
        return; /* Stop before mutating browser history or opening dialogs. */
    }

    currentURL.searchParams.delete('modal'); /* Remove the one-shot profile-modal marker from the URL. */
    window.history.replaceState({}, '', `${currentURL.pathname}${currentURL.search}${currentURL.hash}`); /* Keep the current page clean after consuming the modal marker. */
    await openProfileModal(elements, state); /* Open the profile modal after the dashboard payload finishes loading. */
}

function redirectToLogin(clearTokens) { /* Return to the login route, optionally clearing local tokens first. */
    if (clearTokens) { /* Handle redirects that should clear stale local session state. */
        window.AppWeb.clearTokens(); /* Remove both auth tokens from localStorage. */
    }

    window.goTo('/login'); /* Navigate to the unified login route. */
}

function openDeleteItemConfirmModal(elements, state, itemID, closeItemModalAfterDelete) { /* Stage one item delete and open the custom confirm modal. */
    state.pendingDeleteItemID = itemID; /* Store the staged item ID so the confirm button can use it later. */
    state.closeItemModalAfterDelete = closeItemModalAfterDelete; /* Remember whether the edit modal should close after a confirmed delete. */
    showDialog(elements.deleteItemConfirmModal); /* Open the custom delete-confirm dialog. */
}

function resetDeleteItemConfirmState(state) { /* Clear the staged item-delete confirmation state. */
    state.pendingDeleteItemID = ''; /* Drop the staged item ID after the confirm dialog closes. */
    state.closeItemModalAfterDelete = false; /* Reset the modal-origin delete marker after the confirm dialog closes. */
}

function syncPaginationStateFromInputs(elements, state) { /* Normalize current paging inputs into state. */
    state.pagination.page = Math.max(1, Number(elements.itemsPageInput.value) || 1); /* Normalize the requested page number. */
    state.pagination.limit = Math.max(1, Number(elements.itemsLimitInput.value) || 10); /* Normalize the requested page size. */
    elements.itemsPageInput.value = String(state.pagination.page); /* Mirror the normalized page back into the input. */
    elements.itemsLimitInput.value = String(state.pagination.limit); /* Mirror the normalized limit back into the input. */
}

async function refreshSession(silent, statusNode) { /* Refresh the current access token using the HttpOnly refresh-token cookie. */
    const hasSessionMarker = window.AppWeb.hasSessionMarker(); /* Check whether the browser still has a readable session marker cookie. */

    if (!hasSessionMarker) { /* Guard against refresh attempts when there is no browser-visible session marker. */
        if (!silent && statusNode) { /* Surface missing cookie-backed session state only for visible refresh operations. */
            window.AppWeb.setStatus(statusNode, dashboardT('dashboard.signInAgain', 'Sign in again.'), 'error'); /* Ask the user to authenticate again. */
        }
        return false; /* Stop before calling the refresh endpoint. */
    }

    try { /* Execute the refresh request and handle the result. */
        const result = await window.AppWeb.apiRequest('/auth/refresh', { /* Call the refresh endpoint. */
            method: 'POST', /* Send the request as POST. */
        });

        if (!result.response.ok) { /* Handle failed refresh attempts. */
            if (!silent && statusNode) { /* Surface refresh failures only for visible refresh operations. */
                window.AppWeb.setStatus(statusNode, result.payload?.message || dashboardT('dashboard.signInAgain', 'Sign in again.'), 'error'); /* Surface the refresh failure. */
            }
            return false; /* Tell the caller that the token refresh failed. */
        }

        window.AppWeb.writeTokens({ /* Persist the refreshed token pair into localStorage. */
            accessToken: result.payload?.['access_token'] || '', /* Save the refreshed access token. */
        });
        return true; /* Tell the caller that the refresh succeeded. */
    } catch (_error) { /* Handle network-level failures. */
        if (!silent && statusNode) { /* Surface transport failures only for visible refresh operations. */
            window.AppWeb.setStatus(statusNode, dashboardT('dashboard.networkError', 'Network error. Please try again.'), 'error'); /* Surface a generic transport failure. */
        }
        return false; /* Tell the caller that the refresh failed. */
    }
}

async function runAuthorizedRequest(path, options = {}, statusNode) { /* Execute an authenticated request with one silent refresh retry. */
    let tokens = window.AppWeb.readTokens(); /* Read the currently stored token pair. */
    const hasSessionMarker = window.AppWeb.hasSessionMarker(); /* Check whether the browser still has a readable session marker cookie. */

    if (tokens.accessToken === '' && hasSessionMarker) { /* Refresh first when only cookie-backed refresh state remains. */
        const refreshed = await refreshSession(true, statusNode); /* Attempt a silent refresh before the protected request. */
        if (!refreshed) { /* Handle failed pre-request refresh attempts. */
            return {response: null, payload: null}; /* Signal that the session could not be restored. */
        }

        tokens = window.AppWeb.readTokens(); /* Re-read the stored token pair after refresh. */
    }

    if (tokens.accessToken === '') { /* Guard against protected requests without any access token. */
        return {response: null, payload: null}; /* Signal that the request cannot proceed without authentication. */
    }

    let result = await window.AppWeb.apiRequest(path, { /* Execute the protected request with the current access token. */
        ...options, /* Preserve the caller-provided request options. */
        accessToken: tokens.accessToken, /* Attach the stored access token as a bearer token. */
    });

    if (result.response.status === 401 && window.AppWeb.hasSessionMarker()) { /* Retry once after silently refreshing on unauthorized responses when refresh-cookie state still exists. */
        const refreshed = await refreshSession(true, statusNode); /* Attempt a silent refresh before retrying the request. */

        if (!refreshed) { /* Handle failed silent refresh attempts. */
            return {response: null, payload: null}; /* Signal that the session cannot be recovered. */
        }

        tokens = window.AppWeb.readTokens(); /* Re-read the stored token pair after refresh. */
        result = await window.AppWeb.apiRequest(path, { /* Retry the original request with the refreshed access token. */
            ...options, /* Preserve the caller-provided request options. */
            accessToken: tokens.accessToken, /* Attach the refreshed access token as a bearer token. */
        });
    }

    if (result.response.status === 401) { /* Treat repeated unauthorized responses as invalid sessions. */
        return {response: null, payload: null}; /* Signal that the session is no longer valid. */
    }

    return result; /* Return the final protected request result to the caller. */
}

async function loadCurrentUser(elements, state) { /* Load the authenticated user and hydrate profile forms. */
    const result = await runAuthorizedRequest('/users/me', {method: 'GET'}, elements.dashboardStatus); /* Call the current-user endpoint with auth handling. */

    if (result.response === null) { /* Handle expired or missing sessions uniformly. */
        redirectToLogin(true); /* Return to the login route after clearing stale tokens. */
        return false; /* Tell the caller that no authenticated user could be loaded. */
    }

    if (!result.response.ok) { /* Handle non-auth current-user failures. */
        window.AppWeb.setStatus(elements.dashboardStatus, result.payload?.message || dashboardT('dashboard.failedLoadUser', 'Failed to load current user.'), 'error'); /* Surface the current-user failure. */
        return false; /* Stop before mutating local user state. */
    }

    state.currentUser = result.payload; /* Store the returned user payload in local state. */
    hydrateProfileFields(elements, state.currentUser); /* Mirror the returned user into the profile modal. */
    syncProfileAvatarState(elements, state.currentUser); /* Re-render the profile avatar circle from the current user payload. */
    return true; /* Tell the caller that the current user loaded successfully. */
}

function hydrateProfileFields(elements, user) { /* Mirror a user payload into the profile modal forms. */
    elements.profileNameInput.value = user?.name || ''; /* Fill the profile name field. */
    elements.profileUsernameInput.value = user?.username || ''; /* Fill the profile username field. */
    elements.profileEmailInput.value = user?.email || ''; /* Fill the profile email field. */
}

function syncProfileAvatarState(elements, user) { /* Re-render the account modal avatar state from the current user payload. */
    const avatarURL = typeof user?.['avatar'] === 'string' ? user['avatar'].trim() : ''; /* Normalize the current avatar URL. */

    if (avatarURL !== '') { /* Render the current avatar image when one exists. */
        elements.avatarPlaceholder.hidden = true; /* Hide the avatar placeholder artwork while the current avatar image is visible. */
        elements.avatarPreview.src = avatarURL; /* Point the avatar preview image at the current object URL. */
        elements.avatarPreview.hidden = false; /* Reveal the avatar preview image inside the circle. */
        elements.avatarPickerButton.classList.add('has-image'); /* Switch the avatar circle into image mode. */
        elements.deleteAvatarButton.hidden = false; /* Reveal the delete-avatar button when an avatar exists. */
        elements.avatarPickerButton.title = dashboardT('profile.changeAvatar', 'Change avatar'); /* Update the avatar circle tooltip for replacement uploads. */
        return; /* Stop before applying the empty-avatar state. */
    }

    elements.avatarPlaceholder.hidden = false; /* Reveal the avatar placeholder artwork when no avatar exists. */
    elements.avatarPreview.hidden = true; /* Hide the avatar preview image when no avatar exists. */
    elements.avatarPreview.removeAttribute('src'); /* Clear any stale avatar preview URL. */
    elements.avatarPickerButton.classList.remove('has-image'); /* Restore the empty avatar-circle styling. */
    elements.deleteAvatarButton.hidden = true; /* Hide the delete-avatar button when nothing is uploaded. */
    elements.avatarPickerButton.title = dashboardT('profile.uploadAvatar', 'Upload avatar'); /* Update the avatar circle tooltip for first-time uploads. */
}

function clearItemPreviewObjectURL(state) { /* Revoke any temporary blob URL created for an unsaved item image selection. */
    if (state.itemPreviewObjectURL === '') { /* Skip work when no temporary item-image blob URL exists. */
        return; /* Stop before calling revokeObjectURL unnecessarily. */
    }

    window.URL.revokeObjectURL(state.itemPreviewObjectURL); /* Release the temporary browser blob URL for the last pending item image. */
    state.itemPreviewObjectURL = ''; /* Clear the stored temporary item-image blob URL. */
}

function renderItemImageState(elements, state, imageURL) { /* Re-render the item-image circle from either a persisted or temporary preview URL. */
    if (imageURL !== '') { /* Render an item-image preview whenever any browser-loadable URL exists. */
        elements.itemImagePreview.src = imageURL; /* Point the item-image preview node at the chosen preview URL. */
        elements.itemImagePreview.hidden = false; /* Reveal the item-image preview inside the circle. */
        elements.itemImagePickerButton.classList.add('has-image'); /* Switch the item-image circle into image mode. */
        elements.removeItemImageButton.hidden = false; /* Reveal the item-image delete cross while an image is present. */
        elements.itemImagePickerButton.title = dashboardT('item.changeImage', 'Change image'); /* Keep the tooltip aligned with the visible replace-image affordance. */
        return; /* Stop before applying the empty item-image state. */
    }

    elements.itemImagePreview.hidden = true; /* Hide the item-image preview node when no image exists. */
    elements.itemImagePreview.removeAttribute('src'); /* Clear any stale item-image source URL. */
    elements.itemImagePickerButton.classList.remove('has-image'); /* Restore the empty-circle item-image state. */
    elements.removeItemImageButton.hidden = true; /* Hide the item-image delete cross when nothing is present. */
    elements.itemImagePickerButton.title = dashboardT('item.uploadImage', 'Upload image'); /* Reset the item-image circle tooltip for first-time uploads. */
}

function syncItemImageFromPersistedURL(elements, state, imageURL) { /* Re-render the item-image circle from the persisted item-image URL. */
    state.itemPersistedImageURL = typeof imageURL === 'string' ? imageURL.trim() : ''; /* Normalize the persisted item-image URL before rendering it. */
    clearItemPreviewObjectURL(state); /* Drop any temporary unsaved item-image preview before restoring persisted state. */
    renderItemImageState(elements, state, state.itemPersistedImageURL); /* Render the persisted item-image preview or the empty circle. */
}

function syncItemImagePreviewFromSelection(elements, state) { /* Re-render the item-image circle from the currently selected unsaved file. */
    const selectedFile = elements.itemImageFileInput.files[0]; /* Read the current item-image file selection. */

    if (!selectedFile) { /* Restore persisted item-image state when the file selection becomes empty. */
        syncItemImageFromPersistedURL(elements, state, state.itemPersistedImageURL); /* Fall back to the stored image or the empty circle. */
        return; /* Stop before creating a temporary preview URL. */
    }

    clearItemPreviewObjectURL(state); /* Revoke any older temporary preview before replacing it. */
    state.itemPreviewObjectURL = window.URL.createObjectURL(selectedFile); /* Build a temporary browser URL for the newly selected item image. */
    renderItemImageState(elements, state, state.itemPreviewObjectURL); /* Preview the unsaved item image inside the circle immediately. */
}

async function loadItems(elements, state) { /* Load the current page of items for the authenticated user. */
    const query = new URLSearchParams({ /* Build the paging query string from current state. */
        page: String(state.pagination.page), /* Serialize the current page number. */
        limit: String(state.pagination.limit), /* Serialize the current page size. */
    });

    const result = await runAuthorizedRequest(`/items?${query.toString()}`, {method: 'GET'}, elements.dashboardStatus); /* Call the paginated items endpoint with auth handling. */

    if (result.response === null) { /* Handle expired or missing sessions uniformly. */
        redirectToLogin(true); /* Return to the login route after clearing stale tokens. */
        return false; /* Tell the caller that the item list could not be loaded. */
    }

    if (!result.response.ok) { /* Handle non-auth item-list failures. */
        window.AppWeb.setStatus(elements.dashboardStatus, result.payload?.message || dashboardT('dashboard.failedLoadItems', 'Failed to load items.'), 'error'); /* Surface the item-list failure. */
        return false; /* Stop before mutating local item state. */
    }

    state.listedItems = Array.isArray(result.payload?.['data']) ? result.payload['data'] : []; /* Store the returned item array in local state. */
    state.pagination.totalObjects = Number(result.payload?.['total_objects']) || 0; /* Store the total object count. */
    state.pagination.totalPages = Math.max(1, Number(result.payload?.['total_pages']) || 1); /* Store the total page count. */
    state.pagination.page = Math.max(1, Number(result.payload?.['current_page']) || state.pagination.page); /* Store the authoritative current page. */
    state.pagination.limit = Math.max(1, Number(result.payload?.['limit_per_page']) || state.pagination.limit); /* Store the authoritative page size. */
    elements.itemsPageInput.value = String(state.pagination.page); /* Mirror the current page back into the input. */
    elements.itemsLimitInput.value = String(state.pagination.limit); /* Mirror the current limit back into the input. */
    updatePagingSummary(elements, state); /* Re-render the paging summary line. */
    updatePagingControls(elements, state); /* Enable or disable paging buttons for the current page. */
    renderItemsList(elements, state); /* Re-render the item-card grid. */
    return true; /* Tell the caller that the items loaded successfully. */
}

function updatePagingSummary(elements, state) { /* Re-render the compact paging summary line. */
    elements.itemsSummary.textContent = dashboardFormat('dashboard.itemsSummary', {
        page: state.pagination.page,
        totalPages: state.pagination.totalPages,
        totalObjects: state.pagination.totalObjects
    }, `Page ${state.pagination.page} of ${state.pagination.totalPages} · ${state.pagination.totalObjects} item(s) total`); /* Render the current paging summary. */
}

function updatePagingControls(elements, state) { /* Enable or disable paging controls based on current state. */
    elements.prevPageButton.disabled = state.pagination.page <= 1; /* Disable previous-page on the first page. */
    elements.nextPageButton.disabled = state.pagination.page >= state.pagination.totalPages; /* Disable next-page on the last page. */
    renderPaginationPages(elements, state); /* Re-render the numbered pagination strip for the current page window. */
}

function renderItemsList(elements, state) { /* Render the current page of items into cards. */
    if (state.listedItems.length === 0) { /* Handle empty pages explicitly. */
        elements.itemsList.innerHTML = `<p class="empty-state">${window.AppWeb.escapeHTML(dashboardT('dashboard.noItemsPage', 'No items on this page.'))}</p>`; /* Render a page-level empty state. */
        return; /* Stop before building item-card markup. */
    }

    elements.itemsList.innerHTML = state.listedItems.map((item) => { /* Convert each item payload into a compact card fragment. */
        const rawTitle = item['title'] || dashboardT('item.untitled', 'Untitled item'); /* Resolve the visible item title with a translated fallback. */
        const title = window.AppWeb.escapeHTML(rawTitle); /* Escape the item title for safe rendering. */
        const description = typeof item['description'] === 'string' ? item['description'].trim() : ''; /* Normalize the optional item description before rendering it. */
        const imageURL = window.AppWeb.escapeHTML(item['image'] || ''); /* Escape the stored item image URL when one exists. */
        const itemID = window.AppWeb.escapeHTML(item['id'] || ''); /* Escape the item ID for action attributes. */
        const createdAtValue = formatTimestamp(item['created_at']); /* Normalize the raw creation timestamp once before formatting labels. */
        const updatedAtValue = formatTimestamp(item['updated_at']); /* Normalize the raw update timestamp once before formatting labels. */
        const imageAlt = window.AppWeb.escapeHTML(dashboardFormat('item.imageAlt', {title: rawTitle}, `${rawTitle} image`)); /* Build translated alt text for the optional item thumbnail. */
        const createdLabel = window.AppWeb.escapeHTML(dashboardFormat('item.createdAt', {value: createdAtValue}, `Created ${createdAtValue}`)); /* Build the translated creation caption for the card meta block. */
        const updatedLabel = window.AppWeb.escapeHTML(dashboardFormat('item.updatedAt', {value: updatedAtValue}, `Updated ${updatedAtValue}`)); /* Build the translated update caption for the card meta block. */
        const editLabel = window.AppWeb.escapeHTML(dashboardT('item.editButton', 'Edit')); /* Build the translated edit action label. */
        const deleteLabel = window.AppWeb.escapeHTML(dashboardT('item.deleteButton', 'Delete')); /* Build the translated delete action label. */
        const imageMarkup = imageURL === '' ? '' : `<div class="item-card-media"><img class="item-card-image" src="${imageURL}" alt="${imageAlt}"></div>`; /* Render a thumbnail only when the item has an image. */
        const descriptionMarkup = description === '' ? '' : `<div class="item-card-description">${window.AppWeb.escapeHTML(description)}</div>`; /* Render the description block only when the item actually has one. */

        return `
            <article class="item-card">
                ${imageMarkup}
                <div class="item-card-body">
                    <div class="item-card-title-row">
                        <div class="item-card-title">${title}</div>
                    </div>
                    ${descriptionMarkup}
                </div>
                <div class="item-card-side">
                    <div class="item-card-meta">
                        <span>${createdLabel}</span>
                        <span>${updatedLabel}</span>
                    </div>
                    <div class="item-card-actions">
                        <button class="ghost-btn" type="button" data-action="edit-item" data-item-id="${itemID}">${editLabel}</button>
                        <button class="ghost-btn ghost-btn-danger" type="button" data-action="delete-item" data-item-id="${itemID}">${deleteLabel}</button>
                    </div>
                </div>
            </article>
        `; /* Return the fully rendered item-card HTML. */
    }).join(''); /* Join every rendered card fragment into one HTML string. */
}

function renderPaginationPages(elements, state) { /* Render up to ten numbered page buttons between the arrow controls. */
    const totalPages = Math.max(1, state.pagination.totalPages); /* Normalize the total page count for rendering. */
    const currentPage = Math.min(totalPages, Math.max(1, state.pagination.page)); /* Clamp the current page to the available range. */
    const maxVisiblePages = 10; /* Limit the numbered page strip to ten buttons. */

    let startPage = Math.max(1, currentPage - Math.floor(maxVisiblePages / 2)); /* Start the visible page window near the current page. */
    let endPage = Math.min(totalPages, startPage + maxVisiblePages - 1); /* Cap the visible page window at the last page. */
    startPage = Math.max(1, endPage - maxVisiblePages + 1); /* Backfill the visible page window when near the end. */

    const buttons = []; /* Collect rendered numbered pagination buttons. */

    for (let pageNumber = startPage; pageNumber <= endPage; pageNumber += 1) { /* Build every numbered page button in the visible window. */
        const pageClassName = pageNumber === currentPage ? 'pagination-page is-active' : 'pagination-page'; /* Highlight the active page button. */
        buttons.push(`<button class="${pageClassName}" type="button" data-page-number="${pageNumber}" aria-current="${pageNumber === currentPage ? 'page' : 'false'}">${pageNumber}</button>`); /* Render one numbered page button. */
    }

    elements.itemsPaginationPages.innerHTML = buttons.join(''); /* Replace the numbered pagination strip with the latest button set. */
}

function resetItemModal(elements, state) { /* Reset the item modal into a clean create state. */
    state.itemModalMode = 'create'; /* Switch the modal back into create mode. */
    elements.itemForm.reset(); /* Reset all item form controls. */
    elements.itemIDInput.value = ''; /* Clear the hidden item ID field. */
    elements.itemModalModeLabel.textContent = dashboardT('item.create', 'Create'); /* Reset the item-modal eyebrow label. */
    elements.itemModalTitle.textContent = ''; /* Keep the hidden item-modal title empty. */
    syncItemImageFromPersistedURL(elements, state, ''); /* Reset the item-image circle into its empty state. */
    elements.deleteItemButton.hidden = true; /* Hide item deletion in create mode. */
    window.AppWeb.setStatus(elements.itemModalStatus, '', undefined); /* Clear the item-modal status line. */
}

function populateItemModal(elements, state, item) { /* Mirror an item payload into the edit modal. */
    state.itemModalMode = 'edit'; /* Switch the modal into edit mode. */
    elements.itemForm.reset(); /* Reset stale form values before hydrating the item. */
    elements.itemIDInput.value = item.id || ''; /* Fill the hidden item ID field. */
    elements.itemTitleInput.value = item.title || ''; /* Fill the item title field. */
    elements.itemDescriptionInput.value = item.description || ''; /* Fill the item description field. */
    elements.itemModalModeLabel.textContent = dashboardT('item.edit', 'Edit'); /* Update the item-modal eyebrow label. */
    elements.itemModalTitle.textContent = ''; /* Keep the hidden item-modal title empty. */
    syncItemImageFromPersistedURL(elements, state, item.image || ''); /* Mirror the persisted item-image URL into the preview circle. */
    elements.deleteItemButton.hidden = false; /* Show item deletion in edit mode. */
    window.AppWeb.setStatus(elements.itemModalStatus, '', undefined); /* Clear stale modal status. */
}

function openCreateItemModal(elements, state) { /* Open the item modal in create mode. */
    resetItemModal(elements, state); /* Reset the modal into a clean create state. */
    showDialog(elements.itemModal); /* Open the item modal dialog. */
}

async function openEditItemModal(elements, state, itemID) { /* Load one item and open the modal in edit mode. */
    const result = await runAuthorizedRequest(`/items/${itemID}`, {method: 'GET'}, elements.dashboardStatus); /* Load the target item payload. */

    if (result.response === null) { /* Handle expired or missing sessions uniformly. */
        redirectToLogin(true); /* Return to the login route after clearing stale tokens. */
        return; /* Stop before opening the modal. */
    }

    if (!result.response.ok) { /* Handle non-auth single-item failures. */
        window.AppWeb.setStatus(elements.dashboardStatus, result.payload?.message || dashboardT('item.failedLoad', 'Failed to load item.'), 'error'); /* Surface the single-item failure. */
        return; /* Stop before opening the modal. */
    }

    populateItemModal(elements, state, result.payload); /* Mirror the returned item into the edit modal. */
    showDialog(elements.itemModal); /* Open the item modal dialog. */
}

async function saveItemFromModal(elements, state) { /* Create or update the item currently shown in the modal. */
    const title = elements.itemTitleInput.value.trim(); /* Read the current title field. */
    const description = elements.itemDescriptionInput.value.trim(); /* Read the current description field. */
    const imageFile = elements.itemImageFileInput.files[0]; /* Read the optional image file field. */

    if (title === '') { /* Validate that every item keeps a non-empty title. */
        window.AppWeb.setStatus(elements.itemModalStatus, dashboardT('item.titleRequired', 'Title is required.'), 'error'); /* Surface the missing-title error. */
        return; /* Stop before calling item endpoints. */
    }

    window.AppWeb.setStatus(elements.itemModalStatus, dashboardT('item.saving', 'Saving item...'), 'pending'); /* Surface a pending item-save state. */

    if (state.itemModalMode === 'create') { /* Handle create-mode submissions. */
        const payload = {title: title}; /* Start the create-item payload with the required title field. */

        if (description !== '') { /* Include description only when the field is non-empty. */
            payload.description = description; /* Add the optional description field to the create payload. */
        }

        const createResult = await runAuthorizedRequest('/items', {method: 'POST', body: payload}, elements.itemModalStatus); /* Create the new item first. */

        if (!createResult.response) { /* Handle expired or missing sessions during item creation. */
            redirectToLogin(true); /* Return to the login route after clearing stale tokens. */
            return; /* Stop before continuing item creation flow. */
        }

        if (!createResult.response.ok) { /* Handle non-auth item creation failures. */
            window.AppWeb.setStatus(elements.itemModalStatus, createResult.payload?.message || dashboardT('item.failedCreate', 'Failed to create item.'), 'error'); /* Surface the item creation failure. */
            return; /* Stop before attempting image upload. */
        }

        const createdItemID = createResult.payload?.id || ''; /* Read the newly created item ID from the response payload. */

        if (imageFile && createdItemID !== '') { /* Upload the selected image only after the item exists. */
            const uploaded = await uploadItemImage(elements, state, createdItemID, imageFile); /* Upload the image for the new item. */
            if (!uploaded) { /* Stop when the image upload fails after creation. */
                return; /* Keep the modal open so the user can retry the upload. */
            }
        }

        closeDialog(elements.itemModal); /* Close the item modal after a successful create flow. */
        await loadItems(elements, state); /* Reload the current page of items after creation. */
        window.AppWeb.setStatus(elements.dashboardStatus, '', undefined); /* Keep the global status line clean after a successful create. */
        showDashboardToast(elements, dashboardT('item.createdToast', 'Item created.')); /* Surface successful item creation as a bottom toast. */
        return; /* Stop after the create flow completes successfully. */
    }

    const itemID = elements.itemIDInput.value.trim(); /* Read the current item ID in edit mode. */

    if (itemID === '') { /* Guard against edit-mode saves without an existing item ID. */
        window.AppWeb.setStatus(elements.itemModalStatus, dashboardT('item.idMissing', 'Item ID is missing.'), 'error'); /* Surface the missing-item error. */
        return; /* Stop before calling edit endpoints. */
    }

    const payload = {}; /* Start an empty update payload for edit mode. */
    payload.title = title; /* Always send the edited title field. */
    if (description !== '') { /* Include description when the edit description field is non-empty. */
        payload.description = description; /* Add the description field to the update payload. */
    }

    const updateResult = await runAuthorizedRequest(`/items/${itemID}`, {method: 'PATCH', body: payload}, elements.itemModalStatus); /* Update the current item first. */

    if (!updateResult.response) { /* Handle expired or missing sessions during item update. */
        redirectToLogin(true); /* Return to the login route after clearing stale tokens. */
        return; /* Stop before continuing item edit flow. */
    }

    if (!updateResult.response.ok) { /* Handle non-auth item update failures. */
        window.AppWeb.setStatus(elements.itemModalStatus, updateResult.payload?.message || dashboardT('item.failedUpdate', 'Failed to update item.'), 'error'); /* Surface the item update failure. */
        return; /* Stop before attempting image upload. */
    }

    if (imageFile) { /* Upload the selected image after a successful patch. */
        const uploaded = await uploadItemImage(elements, state, itemID, imageFile); /* Upload the replacement image for the current item. */
        if (!uploaded) { /* Stop when the item image upload fails. */
            return; /* Keep the modal open so the user can retry the upload. */
        }
    }

    closeDialog(elements.itemModal); /* Close the item modal after a successful edit flow. */
    await loadItems(elements, state); /* Reload the current page of items after editing. */
    window.AppWeb.setStatus(elements.dashboardStatus, '', undefined); /* Keep the global status line clean after a successful update. */
    showDashboardToast(elements, dashboardT('item.updatedToast', 'Item updated.')); /* Surface successful item updates as a bottom toast. */
}

async function uploadItemImage(elements, state, itemID, imageFile) { /* Upload a replacement image for an item. */
    const formData = new FormData(); /* Build a multipart body for the item image upload. */
    formData.append('image', imageFile); /* Append the selected image file under the backend field name. */

    const result = await runAuthorizedRequest(`/items/${itemID}/image`, {method: 'PUT', body: formData}, elements.itemModalStatus); /* Call the item-image upload endpoint. */

    if (!result.response) { /* Handle expired or missing sessions during image upload. */
        redirectToLogin(true); /* Return to the login route after clearing stale tokens. */
        return false; /* Tell the caller that the image upload did not complete. */
    }

    if (!result.response.ok) { /* Handle non-auth image upload failures. */
        window.AppWeb.setStatus(elements.itemModalStatus, result.payload?.message || dashboardT('item.failedUploadImage', 'Failed to upload item image.'), 'error'); /* Surface the image upload failure. */
        return false; /* Tell the caller that the image upload did not complete. */
    }

    await refreshItemModal(elements, state, itemID); /* Reload the latest item payload back into the modal. */
    return true; /* Tell the caller that the image upload flow succeeded. */
}

async function removeItemImage(elements, state, itemID) { /* Remove the current image from the modal item. */
    window.AppWeb.setStatus(elements.itemModalStatus, '', undefined); /* Keep the inline item-modal status line empty during image removal. */

    const result = await runAuthorizedRequest(`/items/${itemID}/image`, {method: 'DELETE'}, elements.itemModalStatus); /* Call the item-image delete endpoint. */

    if (!result.response) { /* Handle expired or missing sessions during image removal. */
        redirectToLogin(true); /* Return to the login route after clearing stale tokens. */
        return; /* Stop before mutating modal state. */
    }

    if (!result.response.ok && result.response.status !== 204) { /* Handle non-auth image removal failures. */
        window.AppWeb.setStatus(elements.itemModalStatus, result.payload?.message || dashboardT('item.failedRemoveImage', 'Failed to remove item image.'), 'error'); /* Surface the image-removal failure. */
        return; /* Stop before refreshing the modal item. */
    }

    await refreshItemModal(elements, state, itemID); /* Reload the latest item payload back into the modal. */
    await loadItems(elements, state); /* Reload the current item page after image removal. */
    window.AppWeb.setStatus(elements.itemModalStatus, '', undefined); /* Keep the inline item-modal status line empty after a successful image removal. */
    window.AppWeb.setStatus(elements.dashboardStatus, '', undefined); /* Keep the global status line clean after a successful image removal. */
    showDashboardToast(elements, dashboardT('item.imageRemovedToast', 'Item image removed.')); /* Surface successful image removals as a bottom toast. */
}

async function refreshItemModal(elements, state, itemID) { /* Reload one item and mirror it back into the open modal. */
    const result = await runAuthorizedRequest(`/items/${itemID}`, {method: 'GET'}, elements.itemModalStatus); /* Load the latest item payload from the API. */

    if (!result.response) { /* Handle expired or missing sessions during modal refresh. */
        redirectToLogin(true); /* Return to the login route after clearing stale tokens. */
        return false; /* Tell the caller that the modal item could not be refreshed. */
    }

    if (!result.response.ok) { /* Handle non-auth modal item refresh failures. */
        window.AppWeb.setStatus(elements.itemModalStatus, result.payload?.message || dashboardT('item.failedReload', 'Failed to reload item.'), 'error'); /* Surface the modal refresh failure. */
        return false; /* Tell the caller that the modal item could not be refreshed. */
    }

    populateItemModal(elements, state, result.payload); /* Mirror the latest item payload back into the modal fields. */
    return true; /* Tell the caller that the modal item refreshed successfully. */
}

async function deleteItemByID(elements, state, itemID) { /* Delete one item and refresh the current page afterward. */
    const result = await runAuthorizedRequest(`/items/${itemID}`, {method: 'DELETE'}, elements.dashboardStatus); /* Call the item delete endpoint. */

    if (!result.response) { /* Handle expired or missing sessions during item deletion. */
        redirectToLogin(true); /* Return to the login route after clearing stale tokens. */
        return false; /* Tell the caller that the item deletion did not complete. */
    }

    if (!result.response.ok && result.response.status !== 204) { /* Handle non-auth item deletion failures. */
        window.AppWeb.setStatus(elements.dashboardStatus, result.payload?.message || dashboardT('item.failedDelete', 'Failed to delete item.'), 'error'); /* Surface the item deletion failure. */
        return false; /* Tell the caller that the item deletion did not complete. */
    }

    await loadItems(elements, state); /* Reload the current page of items after deletion. */
    window.AppWeb.setStatus(elements.dashboardStatus, '', undefined); /* Keep the global status line clean after a successful delete. */
    showDashboardToast(elements, dashboardT('item.deletedToast', 'Item deleted.')); /* Surface successful item deletes as a bottom toast. */
    return true; /* Tell the caller that the item deletion completed successfully. */
}

async function openProfileModal(elements, state) { /* Load the latest current user and open the profile dialog. */
    const loaded = await loadCurrentUser(elements, state); /* Refresh the current user before opening the profile modal. */

    if (!loaded) { /* Stop when the current user cannot be loaded. */
        return; /* Stop before opening the profile modal. */
    }

    showDialog(elements.profileModal); /* Open the profile dialog shell. */
}

async function saveProfile(elements, state) { /* PATCH the current user's editable profile fields. */
    const payload = {}; /* Start an empty profile update payload. */
    const nextName = elements.profileNameInput.value.trim(); /* Read the candidate next display name. */
    const nextUsername = elements.profileUsernameInput.value.trim(); /* Read the candidate next username. */
    const nextEmail = elements.profileEmailInput.value.trim(); /* Read the candidate next email. */

    if (nextName !== '' && nextName !== state.currentUser?.name) { /* Include name only when it changed and stays non-empty. */
        payload.name = nextName; /* Add the changed name to the profile payload. */
    }

    if (nextUsername !== '' && nextUsername !== state.currentUser?.username) { /* Include username only when it changed and stays non-empty. */
        payload.username = nextUsername; /* Add the changed username to the profile payload. */
    }

    if (nextEmail !== '' && nextEmail !== (state.currentUser?.email || '')) { /* Include email only when it changed and stays non-empty. */
        payload.email = nextEmail; /* Add the changed email to the profile payload. */
    }

    if (Object.keys(payload).length === 0) { /* Guard against profile submits without actual changes. */
        window.AppWeb.setStatus(elements.profileStatus, dashboardT('profile.changeAtLeastOne', 'Change at least one field first.'), 'error'); /* Surface the empty-profile-update error. */
        return; /* Stop before calling profile endpoints. */
    }

    window.AppWeb.setStatus(elements.profileStatus, dashboardT('profile.saving', 'Saving profile...'), 'pending'); /* Surface a pending profile-save state. */
    const result = await runAuthorizedRequest('/users/me', {method: 'PATCH', body: payload}, elements.profileStatus); /* Call the profile update endpoint. */

    if (!result.response) { /* Handle expired or missing sessions during profile save. */
        redirectToLogin(true); /* Return to the login route after clearing stale tokens. */
        return; /* Stop before mutating profile state. */
    }

    if (!result.response.ok) { /* Handle non-auth profile update failures. */
        window.AppWeb.setStatus(elements.profileStatus, result.payload?.message || dashboardT('profile.failedUpdate', 'Failed to update profile.'), 'error'); /* Surface the profile-update failure. */
        return; /* Stop before mutating local user state. */
    }

    state.currentUser = result.payload; /* Store the refreshed user payload in local state. */
    hydrateProfileFields(elements, state.currentUser); /* Mirror the refreshed user into the profile form. */
    syncProfileAvatarState(elements, state.currentUser); /* Re-render the avatar circle from the refreshed user payload. */
    window.AppWeb.setStatus(elements.profileStatus, '', undefined); /* Keep the inline profile status clear after a successful update. */
    showDashboardToast(elements, dashboardT('profile.updatedToast', 'Profile updated.')); /* Surface a bottom toast for a successful profile update. */
}

async function uploadAvatar(elements, state) { /* Upload a new avatar for the current user. */
    const avatarFile = elements.avatarFileInput.files[0]; /* Read the selected avatar file. */

    if (!avatarFile) { /* Guard against avatar uploads without a selected file. */
        return; /* Stop before calling avatar endpoints. */
    }

    const formData = new FormData(); /* Build a multipart request body for the avatar upload. */
    formData.append('avatar', avatarFile); /* Append the selected avatar file under the backend field name. */
    window.AppWeb.setStatus(elements.profileStatus, dashboardT('profile.uploadingAvatar', 'Uploading avatar...'), 'pending'); /* Surface a pending avatar-upload state in the profile status line. */

    const result = await runAuthorizedRequest('/users/me/avatar', {method: 'PUT', body: formData}, elements.profileStatus); /* Call the avatar upload endpoint. */

    if (!result.response) { /* Handle expired or missing sessions during avatar upload. */
        redirectToLogin(true); /* Return to the login route after clearing stale tokens. */
        return; /* Stop before mutating avatar state. */
    }

    if (!result.response.ok) { /* Handle non-auth avatar upload failures. */
        window.AppWeb.setStatus(elements.profileStatus, result.payload?.message || dashboardT('profile.failedUploadAvatar', 'Failed to upload avatar.'), 'error'); /* Surface the avatar-upload failure. */
        return; /* Stop before mutating local user state. */
    }

    state.currentUser = result.payload; /* Store the refreshed user payload in local state. */
    hydrateProfileFields(elements, state.currentUser); /* Mirror the refreshed user into the profile form. */
    syncProfileAvatarState(elements, state.currentUser); /* Re-render the avatar circle from the refreshed user payload. */
    elements.avatarFileInput.value = ''; /* Clear the hidden avatar file input after success. */
    window.AppWeb.setStatus(elements.profileStatus, '', undefined); /* Keep the inline profile status clear after a successful avatar update. */
    showDashboardToast(elements, dashboardT('profile.avatarUpdatedToast', 'Avatar updated.')); /* Surface a bottom toast for a successful avatar update. */
}

async function deleteAvatar(elements, state) { /* Remove the current avatar from the authenticated user. */
    window.AppWeb.setStatus(elements.profileStatus, '', undefined); /* Keep the inline profile status line empty during avatar removal. */
    const result = await runAuthorizedRequest('/users/me/avatar', {method: 'DELETE'}, elements.profileStatus); /* Call the avatar delete endpoint. */

    if (!result.response) { /* Handle expired or missing sessions during avatar removal. */
        redirectToLogin(true); /* Return to the login route after clearing stale tokens. */
        return; /* Stop before mutating avatar state. */
    }

    if (!result.response.ok && result.response.status !== 204) { /* Handle non-auth avatar removal failures. */
        window.AppWeb.setStatus(elements.profileStatus, result.payload?.message || dashboardT('profile.failedDeleteAvatar', 'Failed to delete avatar.'), 'error'); /* Surface the avatar-removal failure. */
        return; /* Stop before refreshing current user data. */
    }

    await loadCurrentUser(elements, state); /* Reload the latest user payload after avatar removal. */
    elements.avatarFileInput.value = ''; /* Clear the hidden avatar file input after a successful removal. */
    window.AppWeb.setStatus(elements.profileStatus, '', undefined); /* Keep the inline profile status clear after a successful avatar removal. */
    showDashboardToast(elements, dashboardT('profile.avatarRemovedToast', 'Avatar removed.')); /* Surface a bottom toast for a successful avatar removal. */
}

async function logOutAndRedirect(elements) { /* Log out the current session and return to login. */
    let tokens = window.AppWeb.readTokens(); /* Read the currently stored access token. */
    const hasSessionMarker = window.AppWeb.hasSessionMarker(); /* Check whether the browser still has a readable session marker cookie. */

    if (tokens.accessToken === '' && hasSessionMarker) { /* Recover the missing access token first when the refresh cookie is still present. */
        const refreshed = await refreshSession(true, elements.dashboardStatus); /* Try to restore the access token before calling the logout endpoint. */

        if (refreshed) { /* Re-read the restored access token only after a successful refresh. */
            tokens = window.AppWeb.readTokens(); /* Load the fresh access token emitted by the refresh flow. */
        }
    }

    if (tokens.accessToken === '') { /* Handle partial or already-cleared local sessions after the best-effort refresh attempt. */
        redirectToLogin(true); /* Return to login after clearing any stale local state. */
        return; /* Stop before making a network request. */
    }

    try { /* Execute the logout request as a best-effort server-side cleanup. */
        await window.AppWeb.apiRequest('/auth/logout', { /* Call the logout endpoint directly. */
            method: 'POST', /* Send the logout request as POST. */
            accessToken: tokens.accessToken, /* Attach the current access token. */
        });
    } catch (_error) { /* Ignore logout transport failures and still clear local session state. */
        noop(); /* Keep the logout flow strictly best effort. */
    }

    redirectToLogin(true); /* Return to the login route after clearing local session state. */
}

function noop() { /* Provide an explicit no-op helper for intentionally ignored branches. */
    /* Do nothing. */
}

function resolveDashboardToastHost(elements) { /* Pick the DOM host that keeps the shared toast above any open modal backdrop. */
    if (isDialogOpen(elements.profileModal)) { /* Prefer the profile modal while it is open so the toast joins the dialog top layer. */
        return elements.profileModal; /* Render the toast inside the open profile dialog. */
    }

    if (isDialogOpen(elements.itemModal)) { /* Fall back to the item modal while it is open for the same top-layer reason. */
        return elements.itemModal; /* Render the toast inside the open item dialog. */
    }

    return document.body; /* Otherwise keep the toast attached to the document body. */
}

function showDashboardToast(elements, message) { /* Show one transient dashboard toast near the bottom center of the viewport. */
    const toastNode = elements.dashboardToast; /* Read the cached dashboard toast node. */

    if (!toastNode || message.trim() === '') { /* Guard against missing toast nodes and empty messages. */
        return; /* Stop before mutating toast state. */
    }

    if (toastNode.dataset.hideTimerID) { /* Cancel any pending toast hide timer before showing a new message. */
        window.clearTimeout(Number(toastNode.dataset.hideTimerID)); /* Clear the previously scheduled toast hide timeout. */
    }

    const toastHost = resolveDashboardToastHost(elements); /* Choose a toast host that stays above the current backdrop state. */

    if (toastNode.parentElement !== toastHost) { /* Reparent the shared toast node only when its current host is wrong. */
        toastHost.appendChild(toastNode); /* Move the toast into the active top-layer host or back to the page body. */
    }

    toastNode.hidden = false; /* Make the toast visible before animating it in. */
    toastNode.textContent = message; /* Replace the visible toast message text. */
    toastNode.classList.remove('is-visible'); /* Reset toast visibility state before retriggering the entrance animation. */

    requestAnimationFrame(() => { /* Wait one frame so the browser can register the reset state first. */
        toastNode.classList.add('is-visible'); /* Animate the toast into view. */
    });

    const hideTimerID = window.setTimeout(() => { /* Schedule the toast to disappear after a short delay. */
        toastNode.classList.remove('is-visible'); /* Animate the toast out of view. */

        window.setTimeout(() => { /* Wait for the exit transition before hiding the toast node entirely. */
            toastNode.hidden = true; /* Remove the toast from layout and accessibility flow after the exit animation. */
            toastNode.textContent = ''; /* Clear the stale toast message text. */
            delete toastNode.dataset.hideTimerID; /* Remove the stored timer marker after the toast fully closes. */
        }, 220); /* Match the CSS exit transition duration closely. */
    }, 2200); /* Keep the toast visible long enough to read without lingering. */

    toastNode.dataset.hideTimerID = String(hideTimerID); /* Store the current hide timer so a later toast can cancel it. */
}

function hasNativeDialogSupport(dialog) { /* Detect whether the current browser supports the native dialog API. */
    return typeof dialog?.showModal === 'function' && typeof dialog?.close === 'function'; /* Treat showModal and close as the minimum native-dialog contract. */
}

function syncFallbackDialogState() { /* Keep the body scroll-lock class in sync with fallback dialog visibility. */
    const hasOpenFallbackDialog = document.querySelector('.modal[data-dialog-fallback="true"][open]') !== null; /* Check whether any fallback dialog is currently open. */
    document.body.classList.toggle('dialog-open', hasOpenFallbackDialog); /* Lock or unlock body scrolling based on fallback modal state. */
}

function isDialogOpen(dialog) { /* Detect dialog visibility consistently across native and fallback implementations. */
    return Boolean(dialog?.open) || dialog?.hasAttribute('open'); /* Treat either the native open property or the open attribute as visible state. */
}

function showDialog(dialog) { /* Open a dialog only when it is not already open. */
    if (!dialog || isDialogOpen(dialog)) { /* Guard against missing dialogs and duplicate open calls. */
        return; /* Stop before trying to open the dialog. */
    }

    if (hasNativeDialogSupport(dialog)) { /* Prefer the native modal API when the browser supports it. */
        dialog.showModal(); /* Open the dialog using the native modal API. */
        return; /* Stop after the native modal opens. */
    }

    dialog.dataset.dialogFallback = 'true'; /* Mark the dialog as running in fallback mode. */
    dialog.setAttribute('open', 'open'); /* Open the dialog by toggling its standard open attribute. */
    syncFallbackDialogState(); /* Apply body scroll locking for the fallback overlay. */
}

function closeDialog(dialog) { /* Close a dialog only when it is currently open. */
    if (!dialog || !isDialogOpen(dialog)) { /* Guard against missing dialogs and duplicate close calls. */
        return; /* Stop before trying to close the dialog. */
    }

    if (hasNativeDialogSupport(dialog)) { /* Prefer the native close flow when the browser supports it. */
        dialog.close(); /* Close the dialog using the native modal API. */
        return; /* Stop after the native dialog closes. */
    }

    dialog.removeAttribute('open'); /* Close the fallback dialog by removing its open attribute. */
    delete dialog.dataset.dialogFallback; /* Remove the fallback marker from the closed dialog. */
    syncFallbackDialogState(); /* Recompute body scroll locking after the dialog closes. */
    dialog.dispatchEvent(new Event('close')); /* Emit a synthetic close event so existing cleanup handlers still run. */
}

function formatTimestamp(value) { /* Convert ISO timestamps into human-friendly local strings. */
    if (!value) { /* Handle empty or missing timestamps explicitly. */
        return dashboardT('common.unknown', 'Unknown'); /* Return a readable fallback timestamp label. */
    }

    const date = new Date(value); /* Parse the timestamp into a Date instance. */

    if (Number.isNaN(date.getTime())) { /* Guard against invalid timestamps. */
        return dashboardT('common.unknown', 'Unknown'); /* Return the same readable fallback for invalid dates. */
    }

    return date.toLocaleString(); /* Render the timestamp using the user's local browser locale. */
}

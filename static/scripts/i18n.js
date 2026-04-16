(function () { /* Manage persisted language selection and expose translation helpers globally. */
    const STORAGE_KEY = 'go_fx_template.language'; /* Keep the persisted language key stable across the app. */
    const LANGUAGES = { /* Enumerate the supported UI languages. */
        en: 'en', /* English UI copy. */
        ru: 'ru', /* Russian UI copy. */
    };

    const translations = { /* Store all frontend translation strings in one place. */
        en: {
            'page.login.title': 'Log In',
            'page.items.title': 'Items',
            'page.changePassword.title': 'Change Password',
            'page.deleteAccount.title': 'Delete Account',
            'page.resetPassword.title': 'Reset Password',
            'page.verifyEmail.title': 'Email Verification',
            'page.error.title': 'Go REST Template',
            'preferences.language': 'Language',
            'preferences.theme': 'Theme',
            'language.en': 'English',
            'language.ru': 'Russian',
            'theme.light': 'Light',
            'theme.system': 'System',
            'theme.dark': 'Dark',
            'theme.button': 'Theme',
            'auth.login': 'Log In',
            'auth.register': 'Register',
            'auth.username': 'Username',
            'auth.password': 'Password',
            'auth.email': 'Email',
            'auth.name': 'Name',
            'auth.forgotPassword': 'Forgot Password',
            'auth.sendResetLink': 'Send Reset Link',
            'auth.back': 'Back',
            'auth.createAccount': 'Register',
            'auth.submittingResetRequest': 'Submitting reset request...',
            'auth.resetRequestFailed': 'Password reset request failed.',
            'auth.resetRequestSuccess': 'If the account exists, an email will arrive shortly.',
            'auth.loggingIn': 'Logging in...',
            'auth.loginFailed': 'Login failed.',
            'auth.creatingAccount': 'Creating account...',
            'auth.registrationFailed': 'Registration failed.',
            'auth.logout': 'Log Out',
            'auth.networkError': 'Network error. Please try again.',
            'auth.formTitleLogin': 'Log In',
            'auth.formTitleForgotPassword': 'Forgot Password',
            'auth.verifyPending': 'Verifying email...',
            'auth.verifyFailed': 'Email verification failed.',
            'auth.verifySuccess': 'Email verified successfully.',
            'dashboard.items': 'Items',
            'dashboard.actions': 'Dashboard actions',
            'dashboard.newItem': 'New Item',
            'dashboard.page': 'Page',
            'dashboard.limit': 'Limit',
            'dashboard.apply': 'Apply',
            'dashboard.loadingItems': 'Loading items...',
            'dashboard.itemsSummary': 'Page {page} of {totalPages} · {totalObjects} item(s) total',
            'dashboard.noItemsPage': 'No items on this page.',
            'dashboard.previousPage': 'Previous page',
            'dashboard.nextPage': 'Next page',
            'dashboard.signInAgain': 'Sign in again.',
            'dashboard.failedLoadUser': 'Failed to load current user.',
            'dashboard.failedLoadItems': 'Failed to load items.',
            'dashboard.networkError': 'Network error. Please try again.',
            'item.create': 'Create',
            'item.edit': 'Edit',
            'item.title': 'Title',
            'item.description': 'Description',
            'item.save': 'Save',
            'item.delete': 'Delete',
            'item.editButton': 'Edit',
            'item.deleteButton': 'Delete',
            'item.createdAt': 'Created {value}',
            'item.updatedAt': 'Updated {value}',
            'item.untitled': 'Untitled item',
            'item.imageAlt': '{title} image',
            'item.currentImage': 'Current item image',
            'item.uploadImage': 'Upload image',
            'item.changeImage': 'Change image',
            'item.confirmDeleteTitle': 'Delete Item',
            'item.confirmDeleteMessage': 'Delete this item permanently?',
            'item.titleRequired': 'Title is required.',
            'item.saving': 'Saving item...',
            'item.failedLoad': 'Failed to load item.',
            'item.failedCreate': 'Failed to create item.',
            'item.failedUpdate': 'Failed to update item.',
            'item.failedUploadImage': 'Failed to upload item image.',
            'item.failedRemoveImage': 'Failed to remove item image.',
            'item.failedReload': 'Failed to reload item.',
            'item.failedDelete': 'Failed to delete item.',
            'item.idMissing': 'Item ID is missing.',
            'item.createdToast': 'Item created.',
            'item.updatedToast': 'Item updated.',
            'item.imageRemovedToast': 'Item image removed.',
            'item.deletedToast': 'Item deleted.',
            'profile.account': 'Account',
            'profile.name': 'Name',
            'profile.username': 'Username',
            'profile.email': 'Email',
            'profile.update': 'Update Profile',
            'profile.changePassword': 'Change Password',
            'profile.deleteAccount': 'Delete Account',
            'profile.noAvatar': 'No avatar',
            'profile.currentAvatar': 'Current avatar',
            'profile.uploadAvatar': 'Upload avatar',
            'profile.changeAvatar': 'Change avatar',
            'profile.changeAtLeastOne': 'Change at least one field first.',
            'profile.saving': 'Saving profile...',
            'profile.failedUpdate': 'Failed to update profile.',
            'profile.updatedToast': 'Profile updated.',
            'profile.uploadingAvatar': 'Uploading avatar...',
            'profile.failedUploadAvatar': 'Failed to upload avatar.',
            'profile.avatarUpdatedToast': 'Avatar updated.',
            'profile.failedDeleteAvatar': 'Failed to delete avatar.',
            'profile.avatarRemovedToast': 'Avatar removed.',
            'changePassword.title': 'Change Password',
            'changePassword.description': 'Enter your current password and choose a new one.',
            'changePassword.current': 'Current',
            'changePassword.new': 'New',
            'changePassword.submit': 'Change Password',
            'changePassword.fillBoth': 'Fill both password fields.',
            'changePassword.pending': 'Changing password...',
            'changePassword.failed': 'Failed to change password.',
            'deleteAccount.title': 'Delete Account',
            'deleteAccount.description': 'Enter your current password to permanently delete this account.',
            'deleteAccount.current': 'Current',
            'deleteAccount.submit': 'Delete Account',
            'deleteAccount.passwordRequired': 'Current password is required.',
            'deleteAccount.pending': 'Deleting account...',
            'deleteAccount.failed': 'Failed to delete account.',
            'reset.passwordReset': 'Password Reset',
            'reset.error': 'Error',
            'reset.newPassword': 'New password',
            'reset.repeatPassword': 'Repeat password',
            'reset.savePassword': 'Save Password',
            'reset.goHome': 'Go Home',
            'reset.passwordsMismatch': 'Passwords do not match.',
            'reset.saving': 'Saving new password...',
            'reset.failed': 'Failed to reset password.',
            'reset.success': 'Password updated. You can return to the app and sign in.',
            'verify.success': 'Success',
            'verify.error': 'Error',
            'verify.manual': 'Manual verification',
            'verify.token': 'Verification token',
            'verify.submit': 'Verify Email',
            'verify.goHome': 'Go Home',
            'error.defaultTitle': 'Error',
            'error.defaultDescription': 'Something went wrong.',
            'error.goHome': 'Go Home',
            'common.cancel': 'Cancel',
            'common.close': 'Close',
            'common.unknown': 'Unknown',
        },
        ru: {
            'page.login.title': 'Вход',
            'page.items.title': 'Предметы',
            'page.changePassword.title': 'Смена пароля',
            'page.deleteAccount.title': 'Удаление аккаунта',
            'page.resetPassword.title': 'Сброс пароля',
            'page.verifyEmail.title': 'Подтверждение почты',
            'page.error.title': 'Go REST Template',
            'preferences.language': 'Язык',
            'preferences.theme': 'Тема',
            'language.en': 'English',
            'language.ru': 'Русский',
            'theme.light': 'Светлая',
            'theme.system': 'Системная',
            'theme.dark': 'Тёмная',
            'theme.button': 'Тема',
            'auth.login': 'Вход',
            'auth.register': 'Регистрация',
            'auth.username': 'Юзернейм',
            'auth.password': 'Пароль',
            'auth.email': 'Почта',
            'auth.name': 'Имя',
            'auth.forgotPassword': 'Забыли пароль',
            'auth.sendResetLink': 'Отправить ссылку',
            'auth.back': 'Назад',
            'auth.createAccount': 'Зарегистрироваться',
            'auth.submittingResetRequest': 'Отправляем запрос на сброс...',
            'auth.resetRequestFailed': 'Не удалось запросить сброс пароля.',
            'auth.resetRequestSuccess': 'Если аккаунт существует, письмо скоро придёт.',
            'auth.loggingIn': 'Входим...',
            'auth.loginFailed': 'Не удалось войти.',
            'auth.creatingAccount': 'Создаём аккаунт...',
            'auth.registrationFailed': 'Не удалось зарегистрироваться.',
            'auth.logout': 'Выйти',
            'auth.networkError': 'Сетевая ошибка. Попробуйте ещё раз.',
            'auth.formTitleLogin': 'Вход',
            'auth.formTitleForgotPassword': 'Сброс пароля',
            'auth.verifyPending': 'Подтверждаем почту...',
            'auth.verifyFailed': 'Не удалось подтвердить почту.',
            'auth.verifySuccess': 'Почта подтверждена.',
            'dashboard.items': 'Предметы',
            'dashboard.actions': 'Действия панели',
            'dashboard.newItem': 'Новый предмет',
            'dashboard.page': 'Страница',
            'dashboard.limit': 'Лимит',
            'dashboard.apply': 'Применить',
            'dashboard.loadingItems': 'Загружаем предметы...',
            'dashboard.itemsSummary': 'Страница {page} из {totalPages} · всего {totalObjects} предмет(ов)',
            'dashboard.noItemsPage': 'На этой странице нет предметов.',
            'dashboard.previousPage': 'Предыдущая страница',
            'dashboard.nextPage': 'Следующая страница',
            'dashboard.signInAgain': 'Войдите снова.',
            'dashboard.failedLoadUser': 'Не удалось загрузить пользователя.',
            'dashboard.failedLoadItems': 'Не удалось загрузить предметы.',
            'dashboard.networkError': 'Сетевая ошибка. Попробуйте ещё раз.',
            'item.create': 'Создать',
            'item.edit': 'Редактировать',
            'item.title': 'Название',
            'item.description': 'Описание',
            'item.save': 'Сохранить',
            'item.delete': 'Удалить',
            'item.editButton': 'Редактировать',
            'item.deleteButton': 'Удалить',
            'item.createdAt': 'Создан {value}',
            'item.updatedAt': 'Обновлён {value}',
            'item.untitled': 'Без названия',
            'item.imageAlt': 'Изображение {title}',
            'item.currentImage': 'Текущее изображение предмета',
            'item.uploadImage': 'Загрузить изображение',
            'item.changeImage': 'Изменить изображение',
            'item.confirmDeleteTitle': 'Удаление предмета',
            'item.confirmDeleteMessage': 'Удалить этот предмет навсегда?',
            'item.titleRequired': 'Название обязательно.',
            'item.saving': 'Сохраняем предмет...',
            'item.failedLoad': 'Не удалось загрузить предмет.',
            'item.failedCreate': 'Не удалось создать предмет.',
            'item.failedUpdate': 'Не удалось обновить предмет.',
            'item.failedUploadImage': 'Не удалось загрузить изображение предмета.',
            'item.failedRemoveImage': 'Не удалось удалить изображение предмета.',
            'item.failedReload': 'Не удалось заново загрузить предмет.',
            'item.failedDelete': 'Не удалось удалить предмет.',
            'item.idMissing': 'Отсутствует ID предмета.',
            'item.createdToast': 'Предмет создан.',
            'item.updatedToast': 'Предмет обновлён.',
            'item.imageRemovedToast': 'Изображение предмета удалено.',
            'item.deletedToast': 'Предмет удалён.',
            'profile.account': 'Аккаунт',
            'profile.name': 'Имя',
            'profile.username': 'Юзернейм',
            'profile.email': 'Почта',
            'profile.update': 'Обновить профиль',
            'profile.changePassword': 'Сменить пароль',
            'profile.deleteAccount': 'Удалить аккаунт',
            'profile.noAvatar': 'Нет аватара',
            'profile.currentAvatar': 'Текущий аватар',
            'profile.uploadAvatar': 'Загрузить аватар',
            'profile.changeAvatar': 'Изменить аватар',
            'profile.changeAtLeastOne': 'Сначала измените хотя бы одно поле.',
            'profile.saving': 'Сохраняем профиль...',
            'profile.failedUpdate': 'Не удалось обновить профиль.',
            'profile.updatedToast': 'Профиль обновлён.',
            'profile.uploadingAvatar': 'Загружаем аватар...',
            'profile.failedUploadAvatar': 'Не удалось загрузить аватар.',
            'profile.avatarUpdatedToast': 'Аватар обновлён.',
            'profile.failedDeleteAvatar': 'Не удалось удалить аватар.',
            'profile.avatarRemovedToast': 'Аватар удалён.',
            'changePassword.title': 'Смена пароля',
            'changePassword.description': 'Введите текущий пароль и задайте новый.',
            'changePassword.current': 'Текущий',
            'changePassword.new': 'Новый',
            'changePassword.submit': 'Сменить пароль',
            'changePassword.fillBoth': 'Заполните оба поля пароля.',
            'changePassword.pending': 'Меняем пароль...',
            'changePassword.failed': 'Не удалось сменить пароль.',
            'deleteAccount.title': 'Удаление аккаунта',
            'deleteAccount.description': 'Введите текущий пароль, чтобы удалить аккаунт навсегда.',
            'deleteAccount.current': 'Текущий',
            'deleteAccount.submit': 'Удалить аккаунт',
            'deleteAccount.passwordRequired': 'Текущий пароль обязателен.',
            'deleteAccount.pending': 'Удаляем аккаунт...',
            'deleteAccount.failed': 'Не удалось удалить аккаунт.',
            'reset.passwordReset': 'Сброс пароля',
            'reset.error': 'Ошибка',
            'reset.newPassword': 'Новый пароль',
            'reset.repeatPassword': 'Повторите пароль',
            'reset.savePassword': 'Сохранить пароль',
            'reset.goHome': 'На главную',
            'reset.passwordsMismatch': 'Пароли не совпадают.',
            'reset.saving': 'Сохраняем новый пароль...',
            'reset.failed': 'Не удалось сбросить пароль.',
            'reset.success': 'Пароль обновлён. Можно вернуться в приложение и войти.',
            'verify.success': 'Успешно',
            'verify.error': 'Ошибка',
            'verify.manual': 'Ручное подтверждение',
            'verify.token': 'Токен подтверждения',
            'verify.submit': 'Подтвердить почту',
            'verify.goHome': 'На главную',
            'error.defaultTitle': 'Ошибка',
            'error.defaultDescription': 'Что-то пошло не так.',
            'error.goHome': 'На главную',
            'common.cancel': 'Отмена',
            'common.close': 'Закрыть',
            'common.unknown': 'Неизвестно',
        },
    };

    function detectInitialLanguage() { /* Pick the initial UI language from persisted state or browser locale. */
        const savedLanguage = window.localStorage.getItem(STORAGE_KEY); /* Read the persisted language preference when available. */

        if (savedLanguage && Object.prototype.hasOwnProperty.call(LANGUAGES, savedLanguage)) { /* Trust valid persisted language selections first. */
            return savedLanguage; /* Use the persisted language. */
        }

        return navigator.language.toLowerCase().startsWith('ru') ? LANGUAGES.ru : LANGUAGES.en; /* Fall back to browser locale with English as the default. */
    }

    let currentLanguage = detectInitialLanguage(); /* Store the current active language in module state. */

    function t(key, fallback) { /* Resolve one translation key against the current active language. */
        const dictionary = translations[currentLanguage] || translations.en; /* Use the current language dictionary with English fallback. */
        return dictionary[key] || translations.en[key] || fallback || key; /* Resolve the translation with graceful fallback ordering. */
    }

    function format(key, params, fallback) { /* Resolve one translation key and interpolate simple placeholder values. */
        return t(key, fallback).replace(/\{(\w+)}/g, (_match, name) => String(params?.[name] ?? '')); /* Replace each named placeholder with the provided parameter value. */
    }

    function applyTranslations() { /* Push the current language dictionary into the live DOM. */
        document.documentElement.lang = currentLanguage; /* Keep the HTML lang attribute aligned with the active language. */

        document.querySelectorAll('[data-i18n]').forEach((node) => { /* Translate plain-text nodes marked with data-i18n. */
            node.textContent = t(node.dataset.i18n, node.textContent); /* Replace node text with the translated string. */
        });

        document.querySelectorAll('[data-i18n-placeholder]').forEach((node) => { /* Translate placeholder text for form controls. */
            node.setAttribute('placeholder', t(node.dataset.i18nPlaceholder, node.getAttribute('placeholder') || '')); /* Replace placeholder text with the translated string. */
        });

        document.querySelectorAll('[data-i18n-title]').forEach((node) => { /* Translate title attributes for tooltips and hints. */
            node.setAttribute('title', t(node.dataset.i18nTitle, node.getAttribute('title') || '')); /* Replace the title attribute with the translated string. */
        });

        document.querySelectorAll('[data-i18n-aria-label]').forEach((node) => { /* Translate aria-label attributes for assistive technology. */
            node.setAttribute('aria-label', t(node.dataset.i18nAriaLabel, node.getAttribute('aria-label') || '')); /* Replace aria-label text with the translated string. */
        });

        document.querySelectorAll('[data-i18n-alt]').forEach((node) => { /* Translate alt text for non-decorative images. */
            node.setAttribute('alt', t(node.dataset.i18nAlt, node.getAttribute('alt') || '')); /* Replace alt text with the translated string. */
        });
    }

    function setLanguage(nextLanguage) { /* Persist and apply one requested language. */
        if (!Object.prototype.hasOwnProperty.call(LANGUAGES, nextLanguage)) { /* Guard against invalid language requests. */
            return; /* Stop before mutating persisted language state. */
        }

        currentLanguage = nextLanguage; /* Store the new active language in module state. */
        window.localStorage.setItem(STORAGE_KEY, nextLanguage); /* Persist the language selection to localStorage. */
        applyTranslations(); /* Push the new language into the current DOM immediately. */
        document.dispatchEvent(new CustomEvent('app:languagechange', {detail: {language: nextLanguage}})); /* Notify page scripts about the language change. */
    }

    window.AppI18n = { /* Expose translation helpers globally for page scripts. */
        LANGUAGES: LANGUAGES, /* Share the supported language list. */
        getLanguage: () => currentLanguage, /* Share the current active language. */
        t: t, /* Share the raw translation getter. */
        format: format, /* Share the interpolated translation helper. */
        setLanguage: setLanguage, /* Share the persisted language setter. */
    };

    document.addEventListener('DOMContentLoaded', () => { /* Apply translations after the initial DOM has been parsed. */
        applyTranslations(); /* Translate the full page once at startup. */
    });
}());

// ===== API Functions =====

async function fetchSettings() {
    try {
        const response = await fetch('/api/settings');
        if (!response.ok) {
            throw new Error(`Failed to fetch settings: ${response.status}`);
        }
        const settings = await response.json();
        return settings;
    } catch (error) {
        console.error('Error fetching settings:', error);
        showNotification('Failed to load settings', 'error');
        return null;
    }
}

async function saveSettings(settings) {
    try {
        const response = await fetch('/api/settings', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(settings),
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || `Failed to save settings: ${response.status}`);
        }

        return await response.json();
    } catch (error) {
        console.error('Error saving settings:', error);
        throw error;
    }
}

// ===== Validation Functions =====

function validateDiscordWebhookURL(url) {
    // Empty string is valid (disables notifications)
    if (!url || url.trim() === '') {
        return { valid: true, error: '' };
    }

    // Check if it's a valid URL format
    try {
        const urlObj = new URL(url);

        // Must be https
        if (urlObj.protocol !== 'https:') {
            return { valid: false, error: 'Discord webhook URL must use HTTPS' };
        }

        // Must be discord.com domain
        if (urlObj.hostname !== 'discord.com' && urlObj.hostname !== 'discordapp.com') {
            return { valid: false, error: 'URL must be a Discord webhook (discord.com or discordapp.com)' };
        }

        // Must contain /api/webhooks/ in the path
        if (!urlObj.pathname.startsWith('/api/webhooks/')) {
            return { valid: false, error: 'URL must be a valid Discord webhook URL (must contain /api/webhooks/)' };
        }

        return { valid: true, error: '' };
    } catch (e) {
        return { valid: false, error: 'Invalid URL format' };
    }
}

// ===== UI Functions =====

function showNotification(message, type = 'info') {
    const container = document.getElementById('toast-container');

    const icons = {
        success: '<path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />',
        error: '<path stroke-linecap="round" stroke-linejoin="round" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />',
        info: '<path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />',
    };

    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.innerHTML = `
        <svg xmlns="http://www.w3.org/2000/svg" class="toast-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            ${icons[type]}
        </svg>
        <div class="toast-content">
            <div class="toast-message">${message}</div>
        </div>
        <button onclick="this.parentElement.remove()" class="toast-close">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
        </button>
    `;

    container.appendChild(toast);

    // Auto-remove after 5 seconds
    setTimeout(() => {
        toast.remove();
    }, 5000);
}

function displayWebhookError(errorMessage) {
    const errorElement = document.getElementById('webhook-error');
    if (errorMessage) {
        errorElement.textContent = errorMessage;
        errorElement.classList.remove('hidden');
    } else {
        errorElement.textContent = '';
        errorElement.classList.add('hidden');
    }
}

function loadSettingsIntoForm(settings) {
    const webhookInput = document.getElementById('discord-webhook-url');

    if (settings && settings.DiscordWebhookURL) {
        webhookInput.value = settings.DiscordWebhookURL;
    } else {
        webhookInput.value = '';
    }

    // Clear any previous errors
    displayWebhookError('');
}

// ===== Event Handlers =====

async function handleSettingsFormSubmit(event) {
    event.preventDefault();

    const webhookInput = document.getElementById('discord-webhook-url');
    const saveButton = document.getElementById('save-settings-btn');
    const webhookURL = webhookInput.value.trim();

    // Validate webhook URL
    const validation = validateDiscordWebhookURL(webhookURL);
    if (!validation.valid) {
        displayWebhookError(validation.error);
        return;
    }

    // Clear error if validation passed
    displayWebhookError('');

    // Disable save button while saving
    saveButton.disabled = true;
    saveButton.classList.add('opacity-50', 'cursor-not-allowed');

    try {
        const settings = {
            DiscordWebhookURL: webhookURL,
        };

        await saveSettings(settings);

        showNotification('Settings saved successfully!', 'success');
    } catch (error) {
        showNotification(error.message || 'Failed to save settings', 'error');
    } finally {
        // Re-enable save button
        saveButton.disabled = false;
        saveButton.classList.remove('opacity-50', 'cursor-not-allowed');
    }
}

// ===== Initialization =====

async function initializeSettingsPage() {
    // Load current settings
    const settings = await fetchSettings();
    if (settings) {
        loadSettingsIntoForm(settings);
    }

    // Set up form submission
    const form = document.getElementById('settings-form');
    form.addEventListener('submit', handleSettingsFormSubmit);

    // Add real-time validation on input
    const webhookInput = document.getElementById('discord-webhook-url');
    webhookInput.addEventListener('blur', () => {
        const webhookURL = webhookInput.value.trim();
        const validation = validateDiscordWebhookURL(webhookURL);
        if (!validation.valid) {
            displayWebhookError(validation.error);
        } else {
            displayWebhookError('');
        }
    });

    // Clear error when user starts typing
    webhookInput.addEventListener('input', () => {
        displayWebhookError('');
    });
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', initializeSettingsPage);

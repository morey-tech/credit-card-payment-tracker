// Global state
let allCards = [];
let allStatements = [];
let currentEditingCardId = null;
let currentDeletingCardId = null;

// ===== API Functions =====

async function fetchCards() {
    try {
        const response = await fetch('/api/v1/cards');
        if (!response.ok) {
            throw new Error(`Failed to fetch cards: ${response.status}`);
        }
        const cards = await response.json();
        return cards || [];
    } catch (error) {
        console.error('Error fetching cards:', error);
        showNotification('Failed to load credit cards', 'error');
        return [];
    }
}

async function fetchStatements() {
    try {
        const response = await fetch('/api/v1/statements');
        if (!response.ok) {
            throw new Error(`Failed to fetch statements: ${response.status}`);
        }
        const statements = await response.json();
        return statements || [];
    } catch (error) {
        console.error('Error fetching statements:', error);
        return [];
    }
}

async function createCard(cardData) {
    try {
        const response = await fetch('/api/v1/cards', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(cardData),
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || `Failed to create card: ${response.status}`);
        }

        return await response.json();
    } catch (error) {
        console.error('Error creating card:', error);
        throw error;
    }
}

async function updateCard(id, cardData) {
    try {
        const response = await fetch(`/api/v1/cards/${id}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(cardData),
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || `Failed to update card: ${response.status}`);
        }

        return await response.json();
    } catch (error) {
        console.error('Error updating card:', error);
        throw error;
    }
}

async function deleteCard(id) {
    try {
        const response = await fetch(`/api/v1/cards/${id}`, {
            method: 'DELETE',
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || `Failed to delete card: ${response.status}`);
        }

        return true;
    } catch (error) {
        console.error('Error deleting card:', error);
        throw error;
    }
}

// ===== Utility Functions =====

function formatOrdinal(day) {
    const j = day % 10;
    const k = day % 100;
    if (j === 1 && k !== 11) return day + "st";
    if (j === 2 && k !== 12) return day + "nd";
    if (j === 3 && k !== 13) return day + "rd";
    return day + "th";
}

function formatLastFour(digits) {
    return `•••• ${digits}`;
}

function formatCurrency(amount) {
    if (amount === null || amount === undefined || amount === 0) {
        return 'N/A';
    }
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD',
    }).format(amount);
}

function calculateDaysUntilDue(statementDate, dueDate) {
    const statement = new Date(statementDate);
    const due = new Date(dueDate);
    const diffTime = due - statement;
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    return diffDays;
}

function getStatementCountForCard(cardId) {
    return allStatements.filter(s => s.card_id === cardId).length;
}

// ===== UI Rendering Functions =====

function renderCardsTable() {
    const tbody = document.getElementById('cards-table-body');

    if (allCards.length === 0) {
        tbody.innerHTML = `
            <tr>
                <td colspan="6" class="px-6 py-12 text-center text-gray-400">
                    <div class="flex flex-col items-center space-y-3">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                        </svg>
                        <p class="text-lg">No credit cards yet</p>
                        <p class="text-sm text-gray-500">Click "Add Card" to get started</p>
                    </div>
                </td>
            </tr>
        `;
        return;
    }

    // Sort cards alphabetically by name
    const sortedCards = [...allCards].sort((a, b) => a.name.localeCompare(b.name));

    tbody.innerHTML = sortedCards.map(card => `
        <tr class="hover:bg-gray-700/50 transition-colors">
            <td class="px-6 py-4">
                <div class="font-medium text-white">${escapeHtml(card.name)}</div>
            </td>
            <td class="px-6 py-4">
                <div class="font-mono text-gray-300">${formatLastFour(card.last_four)}</div>
            </td>
            <td class="px-6 py-4">
                <div class="text-gray-300">${formatOrdinal(card.statement_day)}</div>
            </td>
            <td class="px-6 py-4">
                <div class="text-gray-300">${card.days_until_due} days</div>
            </td>
            <td class="px-6 py-4">
                <div class="text-gray-300">${formatCurrency(card.credit_limit)}</div>
            </td>
            <td class="px-6 py-4 text-right">
                <div class="flex justify-end space-x-2">
                    <button
                        onclick="openEditCardModal(${card.id})"
                        class="text-blue-400 hover:text-blue-300 font-medium transition-colors p-2"
                        title="Edit card">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                        </svg>
                    </button>
                    <button
                        onclick="openDeleteConfirmation(${card.id})"
                        class="text-red-400 hover:text-red-300 font-medium transition-colors p-2"
                        title="Delete card">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                    </button>
                </div>
            </td>
        </tr>
    `).join('');
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// ===== Modal Management Functions =====

function openAddCardModal() {
    currentEditingCardId = null;

    // Reset form
    document.getElementById('card-form').reset();
    document.getElementById('card-id-input').value = '';
    document.getElementById('modal-title').textContent = 'Add Credit Card';

    // Clear all error messages
    clearFormErrors();

    // Show modal
    document.getElementById('card-modal').classList.remove('hidden');
}

function openEditCardModal(cardId) {
    const card = allCards.find(c => c.id === cardId);
    if (!card) {
        showNotification('Card not found', 'error');
        return;
    }

    currentEditingCardId = cardId;

    // Set form title
    document.getElementById('modal-title').textContent = 'Edit Credit Card';

    // Pre-fill form with card data
    document.getElementById('card-id-input').value = card.id;
    document.getElementById('card-name').value = card.name;
    document.getElementById('last-four').value = card.last_four;

    // For editing, we need to construct example dates based on statement_day and days_until_due
    // Use current month as example
    const today = new Date();
    const year = today.getFullYear();
    const month = String(today.getMonth() + 1).padStart(2, '0');
    const day = String(card.statement_day).padStart(2, '0');
    document.getElementById('statement-date-input').value = `${year}-${month}-${day}`;

    // Calculate due date from statement day and days_until_due
    const statementDate = new Date(`${year}-${month}-${day}`);
    const dueDate = new Date(statementDate);
    dueDate.setDate(dueDate.getDate() + card.days_until_due);
    const dueYear = dueDate.getFullYear();
    const dueMonth = String(dueDate.getMonth() + 1).padStart(2, '0');
    const dueDay = String(dueDate.getDate()).padStart(2, '0');
    document.getElementById('due-date-input').value = `${dueYear}-${dueMonth}-${dueDay}`;

    document.getElementById('credit-limit').value = card.credit_limit || '';

    // Clear errors
    clearFormErrors();

    // Show modal
    document.getElementById('card-modal').classList.remove('hidden');
}

function closeCardModal() {
    document.getElementById('card-modal').classList.add('hidden');
    document.getElementById('card-form').reset();
    currentEditingCardId = null;
    clearFormErrors();
}

function openDeleteConfirmation(cardId) {
    const card = allCards.find(c => c.id === cardId);
    if (!card) {
        showNotification('Card not found', 'error');
        return;
    }

    currentDeletingCardId = cardId;

    // Set card details
    document.getElementById('delete-card-name').textContent = card.name;
    document.getElementById('delete-card-last-four').textContent = formatLastFour(card.last_four);

    // Set statement count
    const statementCount = getStatementCountForCard(cardId);
    document.getElementById('delete-statement-count').textContent =
        `This card has ${statementCount} statement${statementCount !== 1 ? 's' : ''}.`;

    // Show modal
    document.getElementById('delete-modal').classList.remove('hidden');
}

function closeDeleteModal() {
    document.getElementById('delete-modal').classList.add('hidden');
    currentDeletingCardId = null;
}

// ===== Form Validation =====

function validateCardForm() {
    let isValid = true;
    clearFormErrors();

    // Validate card name
    const cardName = document.getElementById('card-name').value.trim();
    if (cardName.length < 2 || cardName.length > 255) {
        showFieldError('card-name-error', 'Card name must be between 2 and 255 characters');
        isValid = false;
    }

    // Validate last four
    const lastFour = document.getElementById('last-four').value.trim();
    if (!/^\d{4}$/.test(lastFour)) {
        showFieldError('last-four-error', 'Last four must be exactly 4 digits');
        isValid = false;
    }

    // Validate last statement date
    const statementDate = document.getElementById('statement-date-input').value;
    if (!statementDate) {
        showFieldError('statement-date-error', 'Last statement date is required');
        isValid = false;
    }

    // Validate last due date
    const dueDate = document.getElementById('due-date-input').value;
    if (!dueDate) {
        showFieldError('due-date-error', 'Last due date is required');
        isValid = false;
    } else if (statementDate && new Date(dueDate) <= new Date(statementDate)) {
        showFieldError('due-date-error', 'Last due date must be after last statement date');
        isValid = false;
    }

    // Validate credit limit (optional, but must be non-negative if provided)
    const creditLimit = document.getElementById('credit-limit').value;
    if (creditLimit && parseFloat(creditLimit) < 0) {
        showFieldError('credit-limit-error', 'Credit limit must be non-negative');
        isValid = false;
    }

    return isValid;
}

function clearFormErrors() {
    const errorElements = document.querySelectorAll('[id$="-error"]');
    errorElements.forEach(el => {
        el.textContent = '';
        el.classList.add('hidden');
    });
}

function showFieldError(elementId, message) {
    const errorElement = document.getElementById(elementId);
    if (errorElement) {
        errorElement.textContent = message;
        errorElement.classList.remove('hidden');
    }
}

// ===== CRUD Operations =====

async function handleCardFormSubmit(event) {
    event.preventDefault();

    if (!validateCardForm()) {
        return;
    }

    // Disable submit button
    const submitBtn = document.getElementById('save-card-btn');
    const originalText = submitBtn.textContent;
    submitBtn.disabled = true;
    submitBtn.textContent = currentEditingCardId ? 'Updating...' : 'Saving...';

    try {
        const cardData = {
            name: document.getElementById('card-name').value.trim(),
            last_four: document.getElementById('last-four').value.trim(),
            statement_date: document.getElementById('statement-date-input').value,
            due_date: document.getElementById('due-date-input').value,
        };

        const creditLimit = document.getElementById('credit-limit').value;
        if (creditLimit) {
            cardData.credit_limit = parseFloat(creditLimit);
        }

        if (currentEditingCardId) {
            // Update existing card
            await updateCard(currentEditingCardId, cardData);
            showNotification('Credit card updated successfully', 'success');
        } else {
            // Create new card
            await createCard(cardData);
            showNotification('Credit card created successfully', 'success');
        }

        // Close modal and refresh table
        closeCardModal();
        await loadAllData();

    } catch (error) {
        showNotification(error.message || 'Failed to save credit card', 'error');
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = originalText;
    }
}

async function confirmDelete() {
    if (!currentDeletingCardId) return;

    const deleteBtn = document.getElementById('confirm-delete-btn');
    const originalText = deleteBtn.textContent;
    deleteBtn.disabled = true;
    deleteBtn.textContent = 'Deleting...';

    try {
        await deleteCard(currentDeletingCardId);
        showNotification('Credit card deleted successfully', 'success');
        closeDeleteModal();
        await loadAllData();
    } catch (error) {
        showNotification(error.message || 'Failed to delete credit card', 'error');
    } finally {
        deleteBtn.disabled = false;
        deleteBtn.textContent = originalText;
    }
}

// ===== Toast Notifications =====

function showNotification(message, type = 'info') {
    const container = document.getElementById('toast-container');

    const colors = {
        success: 'bg-green-600 border-green-500',
        error: 'bg-red-600 border-red-500',
        info: 'bg-blue-600 border-blue-500',
    };

    const icons = {
        success: '<path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />',
        error: '<path stroke-linecap="round" stroke-linejoin="round" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />',
        info: '<path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />',
    };

    const toast = document.createElement('div');
    toast.className = `${colors[type]} border-l-4 p-4 rounded-lg shadow-lg flex items-center space-x-3 min-w-[300px] max-w-md transform transition-all duration-300 translate-x-0`;
    toast.innerHTML = `
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            ${icons[type]}
        </svg>
        <span class="text-white text-sm font-medium flex-1">${escapeHtml(message)}</span>
        <button onclick="this.parentElement.remove()" class="text-white/80 hover:text-white">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
        </button>
    `;

    container.appendChild(toast);

    // Auto-dismiss after 4 seconds
    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transform = 'translateX(400px)';
        setTimeout(() => toast.remove(), 300);
    }, 4000);
}

// ===== Data Loading =====

async function loadAllData() {
    try {
        // Load cards and statements in parallel
        [allCards, allStatements] = await Promise.all([
            fetchCards(),
            fetchStatements(),
        ]);

        // Render the table
        renderCardsTable();
    } catch (error) {
        console.error('Error loading data:', error);
    }
}

// ===== Event Listeners =====

document.addEventListener('DOMContentLoaded', () => {
    // Load initial data
    loadAllData();

    // Form submit handler
    document.getElementById('card-form').addEventListener('submit', handleCardFormSubmit);

    // Close modals on Escape key
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            closeCardModal();
            closeDeleteModal();
        }
    });
});

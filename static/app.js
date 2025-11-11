// API Base URL
const API_BASE = '/api/v1';

// State
let cardsData = [];
let statementsData = [];

// DOM Elements
const upcomingStatementsList = document.getElementById('upcoming-statements-list');
const actionItemsContainer = document.getElementById('action-items-container');
const pendingPaymentsList = document.getElementById('pending-payments-list');
const pendingPlaceholder = document.getElementById('pending-placeholder');
const actionRequiredCard = document.getElementById('action-required-card');

// Pending Payment Cards Container
const pendingPaymentCardsContainer = document.getElementById('pending-payment-cards-container');

// Modal Elements
const modal = document.getElementById('statement-modal');
const statementForm = document.getElementById('statement-form');
const modalCardName = document.getElementById('modal-card-name');
const cardIdInput = document.getElementById('cardIdInput');
const statementAmountInput = document.getElementById('statement-amount');
const statementDateInput = document.getElementById('statement-date');
const officialDueDateInput = document.getElementById('official-due-date');

// Utility Functions
function formatDate(dateString) {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric' }).format(date);
}

function formatCurrency(amount) {
    return `$${parseFloat(amount).toFixed(2)}`;
}

function getNextStatementDate(card) {
    const today = new Date();
    const currentMonth = today.getMonth();
    const currentYear = today.getFullYear();

    let nextStatementDate = new Date(currentYear, currentMonth, card.statement_day);

    if (nextStatementDate < today) {
        nextStatementDate = new Date(currentYear, currentMonth + 1, card.statement_day);
    }

    return nextStatementDate;
}

function calculateRecommendedPaymentDate(dueDate) {
    const date = new Date(dueDate);
    date.setDate(date.getDate() - 7);
    return date;
}

// API Functions
async function fetchCards() {
    try {
        const response = await fetch(`${API_BASE}/cards`);
        if (!response.ok) throw new Error('Failed to fetch cards');
        cardsData = await response.json();
        return cardsData;
    } catch (error) {
        console.error('Error fetching cards:', error);
        return [];
    }
}

async function fetchStatements() {
    try {
        const response = await fetch(`${API_BASE}/statements`);
        if (!response.ok) throw new Error('Failed to fetch statements');
        statementsData = await response.json();
        return statementsData;
    } catch (error) {
        console.error('Error fetching statements:', error);
        return [];
    }
}

async function createStatement(cardId, statementDate, dueDate, amount) {
    try {
        const response = await fetch(`${API_BASE}/statements`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                card_id: cardId,
                statement_date: statementDate,
                due_date: dueDate,
                amount: parseFloat(amount),
                status: 'pending'
            })
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Failed to create statement: ${errorText}`);
        }

        return await response.json();
    } catch (error) {
        console.error('Error creating statement:', error);
        throw error;
    }
}

// UI Rendering Functions
function renderUpcomingStatements(cards) {
    upcomingStatementsList.innerHTML = '';

    if (cards.length === 0) {
        upcomingStatementsList.innerHTML = '<li class="text-sm text-secondary">No cards found</li>';
        return;
    }

    const upcomingCards = cards
        .map(card => ({
            ...card,
            nextStatement: getNextStatementDate(card)
        }))
        .sort((a, b) => a.nextStatement - b.nextStatement)
        .slice(0, 5);

    upcomingCards.forEach(card => {
        const li = document.createElement('li');
        li.className = 'status-list-item';
        li.innerHTML = `
            <span>${card.name}</span>
            <span class="text-secondary">${formatDate(card.nextStatement)}</span>
        `;
        upcomingStatementsList.appendChild(li);
    });
}

function renderActionRequired(cards, statements) {
    actionItemsContainer.innerHTML = '';

    // Find cards that need statement data entry
    const cardsNeedingData = cards.filter(card => {
        // Check if there's a recent statement (within last 32 days)
        const today = new Date();
        const recentStatements = statements.filter(stmt => {
            if (stmt.card_id !== card.id) return false;
            const stmtDate = new Date(stmt.statement_date);
            const daysDiff = (today - stmtDate) / (1000 * 60 * 60 * 24);
            return daysDiff <= 32;
        });
        return recentStatements.length === 0;
    });

    if (cardsNeedingData.length === 0) {
        actionItemsContainer.innerHTML = '<span class="text-sm text-secondary">All caught up!</span>';
        actionRequiredCard.classList.remove('border-primary');
        actionRequiredCard.classList.add('border-gray');
    } else {
        cardsNeedingData.forEach(card => {
            const div = document.createElement('div');
            div.className = 'status-list-item mb-2';
            div.id = `action-card-${card.id}`;
            div.innerHTML = `
                <span class="font-medium">${card.name}</span>
                <button onclick="openModal(${card.id}, '${card.name}')" class="btn btn-primary btn-sm">
                    Enter Statement Data
                </button>
            `;
            actionItemsContainer.appendChild(div);
        });
    }
}

function renderPendingPayments(cards, statements) {
    pendingPaymentsList.innerHTML = '';

    const pendingStatements = statements
        .filter(stmt => stmt.status === 'pending')
        .sort((a, b) => new Date(a.due_date) - new Date(b.due_date));

    if (pendingStatements.length === 0) {
        pendingPaymentsList.innerHTML = '<li class="text-sm text-secondary">No payments scheduled.</li>';
        return;
    }

    pendingStatements.forEach(stmt => {
        const card = cards.find(c => c.id === stmt.card_id);
        if (!card) return;

        const li = document.createElement('li');
        li.className = 'status-list-item';
        li.innerHTML = `
            <span>${card.name}</span>
            <span class="font-medium text-white">${formatCurrency(stmt.amount)}</span>
            <span class="text-secondary">${formatDate(stmt.due_date)}</span>
        `;
        pendingPaymentsList.appendChild(li);
    });
}

function renderPendingPaymentCards(cards, statements) {
    pendingPaymentCardsContainer.innerHTML = '';

    const pendingStatements = statements
        .filter(stmt => stmt.status === 'pending')
        .sort((a, b) => new Date(a.due_date) - new Date(b.due_date));

    if (pendingStatements.length === 0) {
        pendingPaymentCardsContainer.innerHTML = '<div style="padding: 2rem; text-align: center; color: #6b7280;">No pending payments</div>';
        return;
    }

    pendingStatements.forEach(stmt => {
        const card = cards.find(c => c.id === stmt.card_id);
        if (!card) return;

        const recommendedDate = calculateRecommendedPaymentDate(stmt.due_date);

        const section = document.createElement('section');
        section.className = 'card-detail-section';

        // Build the header with optional button
        let headerHTML = `
            <div class="card-detail-header" style="display: flex; justify-content: space-between; align-items: center;">
                <h2 class="card-detail-title">${card.name}</h2>
        `;

        if (!stmt.scheduled_payment_date) {
            headerHTML += `
                <button onclick="openScheduleModal(${stmt.id}, '${card.name}', '${stmt.due_date}')" class="btn btn-primary">
                    Record Payment
                </button>
            `;
        }

        headerHTML += `</div>`;

        // Build the detail grid
        let gridHTML = `<div class="card-detail-grid">`;

        // Statement Amount
        gridHTML += `
            <div class="card-detail-item">
                <span class="card-detail-label">Statement Amount</span>
                <span class="card-detail-value">${formatCurrency(stmt.amount)}</span>
            </div>
        `;

        // Official Due Date
        gridHTML += `
            <div class="card-detail-item">
                <span class="card-detail-label">Official Due Date</span>
                <span class="card-detail-value">${formatDate(stmt.due_date)}</span>
            </div>
        `;

        // Recommended Payment Date
        gridHTML += `
            <div class="card-detail-item highlighted">
                <span class="card-detail-label success">Recommended Payment Date</span>
                <span class="card-detail-value large">${formatDate(recommendedDate)}</span>
            </div>
        `;

        // Scheduled Payment Date (if exists)
        if (stmt.scheduled_payment_date) {
            gridHTML += `
                <div class="card-detail-item highlighted">
                    <span class="card-detail-label success">Scheduled Payment Date</span>
                    <span class="card-detail-value large" style="color: #10b981;">${formatDate(stmt.scheduled_payment_date)}</span>
                </div>
            `;
        }

        gridHTML += `</div>`;

        section.innerHTML = headerHTML + gridHTML;
        pendingPaymentCardsContainer.appendChild(section);
    });
}

// Modal Functions
function openModal(cardId, cardName) {
    modal.classList.remove('hidden');
    modalCardName.textContent = cardName;
    cardIdInput.value = cardId;

    // Set default statement date to today
    const today = new Date().toISOString().split('T')[0];
    statementDateInput.value = today;

    // Reset form fields
    statementAmountInput.value = '';
    officialDueDateInput.value = '';
}

function closeModal() {
    modal.classList.add('hidden');
}

// Form Submission
statementForm.addEventListener('submit', async function(event) {
    event.preventDefault();

    const cardId = parseInt(cardIdInput.value);
    const amount = statementAmountInput.value;
    const statementDate = statementDateInput.value;
    const dueDate = officialDueDateInput.value;

    if (!cardId || !amount || !statementDate || !dueDate) {
        alert('Please fill in all required fields.');
        return;
    }

    try {
        await createStatement(cardId, statementDate, dueDate, amount);

        // Refresh data
        await loadData();

        closeModal();

        // Show success message
        alert('Statement created successfully!');
    } catch (error) {
        alert('Failed to create statement. Please try again.');
    }
});

// Close modal if clicking outside
modal.addEventListener('click', function(event) {
    if (event.target === modal) {
        closeModal();
    }
});

// Initialize
async function loadData() {
    const [cards, statements] = await Promise.all([
        fetchCards(),
        fetchStatements()
    ]);

    renderUpcomingStatements(cards);
    renderActionRequired(cards, statements);
    renderPendingPayments(cards, statements);
    renderPendingPaymentCards(cards, statements);
}

// Schedule Payment Modal Elements
const scheduleModal = document.getElementById('schedule-modal');
const scheduleForm = document.getElementById('schedule-form');
const scheduleStatementId = document.getElementById('scheduleStatementId');
const scheduleDueDate = document.getElementById('scheduleDueDate');
const scheduleCardName = document.getElementById('schedule-card-name');
const scheduleOfficialDueDate = document.getElementById('schedule-official-due-date');
const scheduledPaymentDateInput = document.getElementById('scheduled-payment-date');

// Schedule Payment Functions
function openScheduleModal(statementId, cardName, dueDate) {
    scheduleModal.classList.remove('hidden');
    scheduleStatementId.value = statementId;
    scheduleDueDate.value = dueDate;
    scheduleCardName.textContent = cardName;
    scheduleOfficialDueDate.textContent = formatDate(dueDate);

    // Calculate and set recommended payment date (7 days before due date)
    const recommendedDate = calculateRecommendedPaymentDate(dueDate);
    const recommendedDateStr = recommendedDate.toISOString().split('T')[0];
    scheduledPaymentDateInput.value = recommendedDateStr;
}

function closeScheduleModal() {
    scheduleModal.classList.add('hidden');
}

async function schedulePayment(statementId, scheduledPaymentDate) {
    try {
        const response = await fetch(`${API_BASE}/statements/${statementId}/schedule`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                scheduled_payment_date: scheduledPaymentDate
            })
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Failed to schedule payment: ${errorText}`);
        }

        return await response.json();
    } catch (error) {
        console.error('Error scheduling payment:', error);
        throw error;
    }
}

// Schedule Form Submission
scheduleForm.addEventListener('submit', async function(event) {
    event.preventDefault();

    const statementId = parseInt(scheduleStatementId.value);
    const scheduledDate = scheduledPaymentDateInput.value;

    if (!statementId || !scheduledDate) {
        alert('Please fill in all required fields.');
        return;
    }

    try {
        await schedulePayment(statementId, scheduledDate);

        // Refresh data
        await loadData();

        closeScheduleModal();

        // Show success message
        alert('Payment scheduled successfully!');
    } catch (error) {
        alert('Failed to schedule payment. Please try again.');
    }
});

// Close schedule modal if clicking outside
scheduleModal.addEventListener('click', function(event) {
    if (event.target === scheduleModal) {
        closeScheduleModal();
    }
});

// Load data when page loads
document.addEventListener('DOMContentLoaded', loadData);

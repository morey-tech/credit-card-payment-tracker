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

// Card Detail Elements
const detailCardName = document.getElementById('detail-card-name');
const detailDueDate = document.getElementById('detail-due-date');
const detailStatementAmount = document.getElementById('detail-statement-amount');
const detailRecPaymentContainer = document.getElementById('detail-rec-payment-container');
const detailRecPaymentDate = document.getElementById('detail-rec-payment-date');

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
        upcomingStatementsList.innerHTML = '<li class="text-sm text-gray-400">No cards found</li>';
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
        li.className = 'flex justify-between items-center text-sm';
        li.innerHTML = `
            <span>${card.name}</span>
            <span class="text-gray-400">${formatDate(card.nextStatement)}</span>
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
        actionItemsContainer.innerHTML = '<span class="text-sm text-gray-400">All caught up!</span>';
        actionRequiredCard.classList.remove('border-blue-500');
        actionRequiredCard.classList.add('border-gray-700');
    } else {
        cardsNeedingData.forEach(card => {
            const div = document.createElement('div');
            div.className = 'flex justify-between items-center mb-2';
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
        pendingPaymentsList.innerHTML = '<li class="text-sm text-gray-400">No payments scheduled.</li>';
        return;
    }

    pendingStatements.forEach(stmt => {
        const card = cards.find(c => c.id === stmt.card_id);
        if (!card) return;

        const recommendedDate = calculateRecommendedPaymentDate(stmt.due_date);

        const li = document.createElement('li');
        li.className = 'flex justify-between items-center text-sm';
        li.innerHTML = `
            <span>${card.name}</span>
            <span class="font-medium text-gray-100">${formatCurrency(stmt.amount)}</span>
            <span class="text-gray-400">Pay by ${formatDate(recommendedDate)}</span>
        `;
        pendingPaymentsList.appendChild(li);
    });
}

function showCardDetails(card, latestStatement) {
    detailCardName.textContent = card.name;

    if (latestStatement) {
        detailDueDate.textContent = formatDate(latestStatement.due_date);
        detailStatementAmount.textContent = formatCurrency(latestStatement.amount);

        const recommendedDate = calculateRecommendedPaymentDate(latestStatement.due_date);
        detailRecPaymentDate.textContent = formatDate(recommendedDate);
        detailRecPaymentContainer.classList.remove('hidden');
    } else {
        detailDueDate.textContent = '---';
        detailStatementAmount.textContent = '$ --.--';
        detailRecPaymentContainer.classList.add('hidden');
    }
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

    // Show details for first card with a pending statement, or just first card
    if (cards.length > 0) {
        const firstCardWithStatement = cards.find(card => {
            return statements.some(stmt => stmt.card_id === card.id && stmt.status === 'pending');
        });

        const cardToShow = firstCardWithStatement || cards[0];
        const latestStatement = statements
            .filter(stmt => stmt.card_id === cardToShow.id)
            .sort((a, b) => new Date(b.statement_date) - new Date(a.statement_date))[0];

        showCardDetails(cardToShow, latestStatement);
    }
}

// Load data when page loads
document.addEventListener('DOMContentLoaded', loadData);

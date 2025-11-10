## Project Proposal: Credit Cards

### Project Overview

The "Credit Cards" is an open-source application designed to significantly enhance and automate the process of managing credit card payments for users of budgeting tools like YNAB and online banking platforms such as EQ Bank. The current workflow involves manual tracking of statement release dates, retrieving statement amounts, calculating payment due dates, and then manually entering this information into YNAB and scheduling payments via a banking app. This project aims to centralize and automate these tasks, reducing the risk of missed payments, improving accuracy, and saving valuable time.

The core functionality will revolve around tracking credit card payment cycles, notifying the user when statements are released, extracting key payment information (amount, due date), and presenting this data in a clear, actionable format. Future iterations will explore direct integration with YNAB and banking APIs to further automate data entry and payment scheduling.

### Key Pain Points Addressed:

1.  **Manual Due Date Tracking:** Inconsistent knowledge of actual payment due dates.
2.  **Statement Amount Retrieval:** Manual checking of credit card accounts for statement amounts.
3.  **YNAB Synchronization:** Manually updating YNAB transactions with correct amounts and dates.
4.  **Payment Scheduling:** Manual scheduling of bill payments via banking apps.
5.  **Lack of Proactive Notifications:** No automated alerts for new statements or upcoming payments.

### Project Goals (Proof of Concept - PoC)

The initial PoC will focus on providing a centralized dashboard for credit card information and robust notification capabilities.

* **Centralized Information Display:** Show critical payment information for all tracked credit cards.
* **Proactive Notifications:** Alert the user via Discord about new statements and upcoming payment recommendations.
* **Guided Workflow:** Provide calculated payment dates and amounts based on user input.

### User Stories (Prioritized for PoC)

The following user stories are ordered to reflect a staged development approach, starting with core information display and notifications, and moving towards more complex integrations.

---

#### Phase 1: Core Information Display & Notifications (PoC)

**User Story 1: Manual Credit Card Configuration**
* **As a user,** I want to be able to manually add and configure my credit cards (e.g., card name, typical statement release day, payment due day relative to statement release) so that the application knows which cards to track.
* **Acceptance Criteria:**
    * The application allows adding a new credit card with a descriptive name.
    * The application allows specifying the typical statement release date (e.g., "3rd of the month", "25th of the month").
    * The application allows specifying the typical number of days between statement release and payment due date.
    * Configured credit cards are displayed in a list.

**User Story 2: Manual Statement Information Input**
* **As a user,** when a new statement is released, I want to be able to manually input the statement amount and the official due date for a specific credit card so that the system has accurate, up-to-date payment information.
* **Acceptance Criteria:**
    * For each configured credit card, there is an option to "Enter New Statement Data."
    * The input form captures the statement amount and the exact due date.
    * Once submitted, this information is associated with the credit card.

**User Story 3: Calculated Payment Recommendation Display**
* **As a user,** after entering new statement information, I want the application to immediately display the recommended payment amount (statement amount) and the recommended payment scheduling date (one week before the official due date) so that I have the exact figures for YNAB and EQ Bank.
* **Acceptance Criteria:**
    * Upon successful input of statement data, the credit card's display updates to show:
        * Statement Amount: \[Inputted Amount]
        * Official Due Date: \[Inputted Date]
        * Recommended Payment Date (for EQ Bank/YNAB): \[Official Due Date - 7 days]

**User Story 4: Discord Statement Release Notification**
* **As a user,** I want to receive a Discord notification when the application predicts a new credit card statement has been released (based on its configuration) so that I am prompted to check my bank and input the statement details.
* **Acceptance Criteria:**
    * The application allows configuration of a Discord webhook URL.
    * On the predicted statement release day for a credit card, a message is sent to the configured Discord channel.
    * The Discord message includes the credit card name and a prompt to log in and get the statement details.
    * *(Optional but desired for PoC)* The Discord message provides a direct link within the application to input the statement details.

**User Story 5: Discord Payment Reminder Notification**
* **As a user,** I want to receive a Discord notification one week before the *official* credit card payment due date (i.e., on the recommended payment date) so that I am reminded to schedule the payment if I haven't already.
* **Acceptance Criteria:**
    * A Discord message is sent on the calculated "Recommended Payment Date."
    * The message includes the credit card name, the statement amount, the official due date, and explicitly states the recommended payment date for scheduling.

---

#### Phase 2: Enhanced Features (Future Roadmap)

**User Story 6: Automated Statement Retrieval (Future)**
* **As a user,** I want the application to securely connect to my credit card provider's online portal (via Plaid/API/screen scraping) and automatically fetch the statement amount and due date when a new statement is released so that I don't have to manually input this information.
* **Acceptance Criteria:**
    * Integration with a financial data aggregator or direct bank APIs.
    * Automatic population of statement amount and due date fields.

**User Story 7: YNAB Transaction Creation (Future)**
* **As a user,** I want the application to integrate with the YNAB API to automatically create or update a scheduled credit card payment transaction in my YNAB budget with the correct amount and recommended payment date so that I don't have to manually enter it.
* **Acceptance Criteria:**
    * Authentication with the YNAB API.
    * Creation/update of transactions in the specified credit card and checking accounts in YNAB.

**User Story 8: EQ Bank Payment Scheduling (Future)**
* **As a user,** I want the application to securely connect to my EQ Bank account (via API/screen scraping) and automatically schedule the bill payment to my credit card with the correct amount and recommended payment date so that I don't have to manually initiate the payment.
* **Acceptance Criteria:**
    * Integration with EQ Bank's API or a secure automation method.
    * Confirmation of scheduled payment within EQ Bank.

**User Story 9: Payment History Tracking (Future)**
* **As a user,** I want the application to track and display a history of my credit card payments (amount, date paid, statement period) so that I can easily review past transactions.
* **Acceptance Criteria:**
    * A historical log view for each credit card.

---

### Technologies (Initial Thoughts for Go/JavaScript)

* **Backend (Go):** Ideal for scheduled tasks (checking for statement dates), Discord webhook integration, and API management (future bank integrations). Robust for handling financial data.
* **Frontend (JavaScript/TypeScript - e.g., React, Vue, Svelte):** For the web-based user interface, allowing manual configuration and data input.
* **Database:** A simple embedded database (e.g., SQLite) for storing credit card configurations and statement history initially, scaling to PostgreSQL/MySQL if needed.
* **Scheduling:** Cron jobs or Go's built-in `time` package for managing daily checks for statement releases and payment reminders.
* **Notifications:** Discord Webhooks for simple and effective alerts.

---

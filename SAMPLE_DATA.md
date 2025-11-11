# Sample Data Documentation

## Overview

The application includes comprehensive sample data to support development and manual testing. The sample data covers various scenarios and edge cases to help validate all features of the Credit Card Payment Tracker.

## Loading Sample Data

Sample data is automatically loaded when the `LOAD_SAMPLE_DATA` environment variable is set to `true`:

```bash
export LOAD_SAMPLE_DATA=true
go run cmd/server/main.go
```

The sample data loading is **idempotent** - running it multiple times will not create duplicates.

---

## Sample Credit Cards (6 cards)

### 1. TD Aeroplan Visa
- **Last Four:** 9876
- **Statement Day:** 15th of each month
- **Days Until Due:** 25 days
- **Credit Limit:** $5,000.00
- **Purpose:** Standard card with medium limit
- **Statements:** 2 (1 paid, 1 pending)

**Statements:**
- Past statement (paid): $1,250.75
- Current statement (pending): $892.50

---

### 2. Amex Cobalt
- **Last Four:** 1234
- **Statement Day:** 28th of each month
- **Days Until Due:** 25 days
- **Credit Limit:** $10,000.00
- **Purpose:** High-limit card with end-of-month cycle
- **Statements:** 2 (1 paid, 1 pending)

**Statements:**
- Past statement (paid): $2,150.00
- Current statement (pending): $3,421.89

---

### 3. Chase Sapphire Reserve
- **Last Four:** 5678
- **Statement Day:** 1st of each month
- **Days Until Due:** 21 days
- **Credit Limit:** *Not set* (testing optional field)
- **Purpose:** Testing card without credit limit + beginning-of-month cycle
- **Statements:** 1 (overdue pending)

**Statements:**
- Overdue statement (pending): $567.25
  - Statement Date: 2 months ago
  - Due Date: 1 month ago (overdue)
  - **Use case:** Testing overdue payment scenarios

---

### 4. Capital One Quicksilver
- **Last Four:** 4321
- **Statement Day:** 5th of each month
- **Days Until Due:** 15 days (short cycle)
- **Credit Limit:** $3,000.00
- **Purpose:** Testing shorter payment cycle + multiple statement history
- **Statements:** 3 (2 paid, 1 pending)

**Statements:**
- Old paid statement: $125.50 (3 months ago)
- Recent paid statement: $435.99 (2 months ago)
- Current pending statement: $15.00 (small amount)
  - **Use case:** Testing small payment amounts

---

### 5. Discover It
- **Last Four:** 8888
- **Statement Day:** 20th of each month
- **Days Until Due:** 30 days (long cycle)
- **Credit Limit:** $7,500.00
- **Purpose:** Testing longer payment cycle + large amounts
- **Statements:** 1 (1 pending with large amount)

**Statements:**
- Current statement (pending): $4,567.89
  - **Use case:** Testing large payment amounts

---

### 6. Citi Double Cash
- **Last Four:** 2468
- **Statement Day:** 10th of each month
- **Days Until Due:** 25 days
- **Credit Limit:** $8,000.00
- **Purpose:** Testing card with NO statements
- **Statements:** 0

**Use case:** Testing empty state - card exists but no statements generated yet

---

## Sample Statements (9 statements total)

### By Status
- **Paid:** 4 statements
- **Pending:** 5 statements
  - 1 overdue (Chase)
  - 4 current/upcoming

### By Amount Range
- **Small amount:** $15.00 (Capital One - testing small payments)
- **Medium amounts:** $125.50 - $1,250.75
- **Large amounts:** $2,150.00 - $4,567.89 (testing formatting with large numbers)

### By Due Date Scenarios
- **Overdue:** 1 statement (Chase - due last month)
- **Current:** 4 statements (various due dates this/next month)
- **Paid (historical):** 4 statements (completed payments)

---

## Test Coverage Scenarios

The sample data is designed to test the following scenarios:

### 1. Credit Card Variations
✅ **Optional field (credit limit):** Chase has no credit limit set
✅ **Various statement days:** 1st, 5th, 10th, 15th, 20th, 28th
✅ **Different due date offsets:** 15, 21, 25, 30 days
✅ **Credit limit range:** $3,000 - $10,000 (plus one with no limit)
✅ **Card with no statements:** Citi (empty state testing)

### 2. Statement Variations
✅ **Paid statements:** Historical payment data
✅ **Pending statements:** Current amounts due
✅ **Overdue statements:** Past-due scenarios
✅ **Amount variations:** Small ($15), medium ($125-$1,250), large ($2,150-$4,567)
✅ **Multiple statements per card:** Capital One has 3 statements
✅ **Notification states:** Various combinations of notified_statement and notified_payment flags

### 3. Dashboard Testing
✅ **Upcoming statements:** Multiple cards with different statement dates
✅ **Action required:** Overdue payment (Chase)
✅ **Pending payments:** Multiple cards with amounts due
✅ **Empty states:** Citi has no statements to display

### 4. Date Calculation Testing
✅ **Short cycle:** 15 days (Capital One)
✅ **Standard cycle:** 21-25 days (most cards)
✅ **Long cycle:** 30 days (Discover)
✅ **Month boundaries:** 1st and 28th statement dates

### 5. Filtering & Sorting
✅ **By status:** Mix of paid/pending
✅ **By due date:** Various dates for sorting tests
✅ **By amount:** Range of amounts for filtering
✅ **By card:** Multiple cards to test grouping

---

## Manual Testing Usage

When performing manual testing using [MANUAL_TESTING_CHECKLIST.md](MANUAL_TESTING_CHECKLIST.md):

1. **Start with sample data loaded** to have realistic test data
2. **Test CRUD operations** alongside existing sample cards
3. **Verify dashboard displays** with multiple cards and statements
4. **Test edge cases:**
   - Delete Capital One (has multiple statements - tests CASCADE)
   - Edit Chase (has no credit limit - tests optional field handling)
   - Mark Chase overdue statement as paid
   - Add new statement to Citi (currently has none)

---

## Database Impact

**Total Records:**
- Credit Cards: 6
- Statements: 9
- Foreign Key Relationships: All 9 statements properly reference valid cards

**Data Integrity:**
- All amounts are positive and realistic ($15.00 - $4,567.89)
- All dates are calculated relative to current date
- All foreign key constraints are satisfied
- All required fields are populated

---

## Development Tips

### Resetting Sample Data

To reload sample data from scratch:

```bash
# Delete the database
rm payment_tracker.db

# Restart with sample data
export LOAD_SAMPLE_DATA=true
go run cmd/server/main.go
```

### Adding More Sample Data

To add additional sample data, edit `pkg/database/sample_data.go`:

1. Add new card insertions before the `Commit()`
2. Add corresponding statements
3. Update the log message with new counts
4. Update `sample_data_test.go` to reflect new counts
5. Run tests: `go test ./pkg/database -v`

---

## Idempotency

The `LoadSampleData` function checks if sample data already exists before inserting:

```go
SELECT COUNT(*) FROM credit_cards
WHERE name IN ('TD Aeroplan Visa', 'Amex Cobalt', ...)
```

If any of the sample cards exist, the function skips loading to prevent duplicates.

---

## Test Data vs Production

**⚠️ IMPORTANT:** Sample data is for development and testing only.

- DO NOT load sample data in production
- Sample data includes fictional credit card information
- The `LOAD_SAMPLE_DATA` environment variable should NOT be set in production deployments

---

## Summary Table

| Card Name | Last 4 | Stmt Day | Days Due | Credit Limit | Statements | Notes |
|-----------|--------|----------|----------|--------------|------------|-------|
| TD Aeroplan Visa | 9876 | 15th | 25 | $5,000 | 2 | Standard card |
| Amex Cobalt | 1234 | 28th | 25 | $10,000 | 2 | High limit, end-of-month |
| Chase Sapphire Reserve | 5678 | 1st | 21 | *None* | 1 | No limit set, overdue stmt |
| Capital One Quicksilver | 4321 | 5th | 15 | $3,000 | 3 | Short cycle, small amount |
| Discover It | 8888 | 20th | 30 | $7,500 | 1 | Long cycle, large amount |
| Citi Double Cash | 2468 | 10th | 25 | $8,000 | 0 | No statements |

**Total:** 6 cards, 9 statements (4 paid, 5 pending)

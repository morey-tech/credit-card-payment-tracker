# Manual Testing Checklist - Phase 10

## Purpose
Comprehensive manual testing checklist for User Story 1: Credit Card Payment Tracker functionality.

---

## Prerequisites
- [x] Application is running locally or accessible
- [ ] Database is initialized
- [ ] Browser DevTools console is open (F12)
- [ ] Test with multiple browsers (Chrome, Firefox, Safari if possible)

---

## 1. Cards Page Testing (`/static/cards.html`)

### 1.1 Add New Card
- [ ] Click "Add Card" button
- [ ] Modal opens with empty form
- [ ] Test field validation:
  - [ ] Submit with empty name → error message shown
  - [ ] Enter name with 1 character → error shown
  - [ ] Enter valid name (2+ characters) → error cleared
  - [ ] Submit with empty last four digits → error shown
  - [ ] Enter 3 digits in last four → error shown
  - [ ] Enter non-numeric last four → error shown
  - [ ] Enter valid 4 digits → error cleared
  - [ ] Submit with empty statement date → error shown
  - [ ] Submit with empty due date → error shown
  - [ ] Enter due date same as statement date → error shown
  - [ ] Enter due date before statement date → error shown
  - [ ] Enter valid dates (due > statement) → error cleared
  - [ ] Enter negative credit limit → error shown
- [ ] Fill all required fields with valid data
- [ ] Click "Add Card"
- [ ] Success notification appears
- [ ] Modal closes
- [ ] New card appears in table
- [ ] Card displays correct information (name, last 4, dates, limit)

### 1.2 Edit Existing Card
- [ ] Click "Edit" button on a card
- [ ] Modal opens with pre-filled form
- [ ] Verify all fields show correct current values
- [ ] Change name to a new valid value
- [ ] Click "Update Card"
- [ ] Success notification appears
- [ ] Modal closes
- [ ] Card table updates with new name
- [ ] Test editing other fields individually:
  - [ ] Last four digits
  - [ ] Statement date (verify due date validation)
  - [ ] Due date (verify it must be after statement date)
  - [ ] Credit limit
- [ ] Test validation errors in edit mode (same as add)

### 1.3 Delete Card
- [ ] Click "Delete" button on a card
- [ ] Confirmation modal appears
- [ ] Verify card name and last four shown in confirmation
- [ ] Click "Cancel" → modal closes, card not deleted
- [ ] Click "Delete" again
- [ ] Click "Confirm Delete"
- [ ] Success notification appears
- [ ] Card removed from table
- [ ] Associated statements are also deleted (verify in statements table if exists)

### 1.4 View Cards Table
- [ ] Table displays all cards
- [ ] Each row shows: Name, Last 4, Statement Day, Due Day offset, Credit Limit, Actions
- [ ] Statement day shown as ordinal (1st, 2nd, 3rd, etc.)
- [ ] Credit limit formatted with commas and 2 decimals
- [ ] Empty state shown when no cards exist

### 1.5 Date Calculations
- [ ] Add card with statement date "2024-11-15" and due date "2024-12-10"
- [ ] Verify "Days Until Due" shows 25
- [ ] Edit card and change dates
- [ ] Verify days recalculated correctly

---

## 2. Settings Page Testing (`/static/settings.html`)

### 2.1 View Settings
- [ ] Navigate to Settings page
- [ ] Discord webhook URL field is displayed
- [ ] Current webhook URL (if any) is shown

### 2.2 Update Discord Webhook
- [ ] Enter invalid URL (e.g., "http://example.com") → error shown
- [ ] Enter valid discord.com webhook URL
- [ ] Click "Save Settings"
- [ ] Success notification appears
- [ ] Reload page → webhook URL persists
- [ ] Enter valid discordapp.com webhook URL → also accepted
- [ ] Leave webhook URL empty → allowed (optional field)
- [ ] Save empty webhook → succeeds

### 2.3 Validation
- [ ] Test invalid webhook formats:
  - [ ] `http://discord.com/api/webhooks/123/abc` → error (http not https)
  - [ ] `https://example.com/webhooks/123/abc` → error (wrong domain)
  - [ ] `https://discord.com/invalid/path` → error (wrong path)
- [ ] Verify error messages are clear and helpful

---

## 3. Dashboard Testing (`/static/index.html`)

### 3.1 Upcoming Statements Section
- [ ] Add multiple cards
- [ ] Verify upcoming statements are displayed
- [ ] Check statement dates are correctly calculated
- [ ] Verify cards are sorted by statement date (soonest first)
- [ ] Empty state shown when no upcoming statements

### 3.2 Action Required Section
- [ ] Create statements with "pending" status
- [ ] Verify they appear in Action Required
- [ ] Verify due dates are shown
- [ ] Empty state when no action required

### 3.3 Pending Payments Section
- [ ] Verify pending payments are displayed
- [ ] Check amounts are formatted correctly
- [ ] Empty state when no pending payments

### 3.4 Navigation
- [ ] All links in dashboard work correctly
- [ ] Links go to correct pages (cards, settings)

---

## 4. Navigation Testing

### 4.1 Sidebar Navigation
- [ ] Sidebar visible on all pages
- [ ] Home icon/link navigates to dashboard
- [ ] Cards icon/link navigates to cards page
- [ ] Settings icon/link navigates to settings page
- [ ] Active page is visually indicated in sidebar
- [ ] Sidebar remains functional across all pages

### 4.2 Page Loading
- [ ] All pages load without JavaScript errors (check console)
- [ ] All pages load without CSS errors
- [ ] All API calls complete successfully (check Network tab)

---

## 5. Responsive Design Testing

### 5.1 Desktop (1920x1080)
- [ ] All pages render correctly
- [ ] Tables are readable
- [ ] Modals are properly centered
- [ ] No horizontal scroll

### 5.2 Laptop (1366x768)
- [ ] Layout adjusts appropriately
- [ ] No content clipping
- [ ] Modals remain accessible

### 5.3 Tablet (768x1024)
- [ ] Sidebar collapses or adapts
- [ ] Tables scroll horizontally if needed
- [ ] Modals fit screen
- [ ] Touch targets are large enough

### 5.4 Mobile (375x667)
- [ ] Content is readable
- [ ] Forms are usable
- [ ] Buttons are tappable
- [ ] No content overflow

---

## 6. Form Validation & Error Handling

### 6.1 Frontend Validation
- [ ] All required field validations work
- [ ] Error messages are clear and helpful
- [ ] Errors clear when fixed
- [ ] Form doesn't submit with validation errors

### 6.2 Backend Error Handling
- [ ] Stop the server
- [ ] Try to add a card → friendly error message shown
- [ ] Restart server
- [ ] Create duplicate card (same name/last four) → handled gracefully
- [ ] Network errors show user-friendly messages

### 6.3 Edge Cases
- [ ] Very long card names (100+ characters) → handled
- [ ] Special characters in card names → handled
- [ ] Large credit limits (999999999.99) → formatted correctly
- [ ] Zero credit limit → allowed
- [ ] Dates in past → allowed (for historical data)

---

## 7. Data Persistence

### 7.1 Database Persistence
- [ ] Add a card
- [ ] Restart the server
- [ ] Verify card still exists
- [ ] Edit the card
- [ ] Restart server
- [ ] Verify edits persisted
- [ ] Delete the card
- [ ] Restart server
- [ ] Verify card is gone

### 7.2 Configuration Persistence
- [ ] Set Discord webhook
- [ ] Restart server
- [ ] Verify webhook persisted in config file
- [ ] Check `config.yaml` file contains correct URL

---

## 8. Accessibility Testing

### 8.1 Keyboard Navigation
- [ ] Tab through all form fields in logical order
- [ ] Enter key submits forms
- [ ] Escape key closes modals
- [ ] All interactive elements reachable by keyboard
- [ ] Focus indicators are visible

### 8.2 Screen Reader Compatibility
- [ ] Form labels are read correctly
- [ ] Error messages are announced
- [ ] Button purposes are clear
- [ ] Table headers are properly associated

### 8.3 Visual Accessibility
- [ ] Color contrast is sufficient (text readable)
- [ ] Focus indicators have good contrast
- [ ] No information conveyed by color alone
- [ ] Text is resizable without breaking layout

---

## 9. Performance Testing

### 9.1 Page Load Times
- [ ] Dashboard loads in < 2 seconds
- [ ] Cards page loads in < 2 seconds
- [ ] Settings page loads in < 2 seconds

### 9.2 API Response Times
- [ ] GET /api/v1/cards responds in < 500ms
- [ ] POST /api/v1/cards responds in < 500ms
- [ ] PUT /api/v1/cards/:id responds in < 500ms
- [ ] DELETE /api/v1/cards/:id responds in < 500ms
- [ ] GET /api/settings responds in < 500ms

### 9.3 Large Data Sets
- [ ] Add 50+ cards
- [ ] Table still renders smoothly
- [ ] Scrolling is smooth
- [ ] No memory leaks (check DevTools Memory tab)

---

## 10. Cross-Browser Testing

### 10.1 Chrome/Edge (Chromium)
- [ ] All functionality works
- [ ] No console errors
- [ ] Styling is correct

### 10.2 Firefox
- [ ] All functionality works
- [ ] No console errors
- [ ] Styling is correct

### 10.3 Safari (if available)
- [ ] All functionality works
- [ ] No console errors
- [ ] Styling is correct

---

## 11. Integration Testing

### 11.1 End-to-End Workflows
- [ ] Complete workflow: Add card → View on dashboard → Edit card → Delete card
- [ ] Multiple cards workflow: Add 3 cards → Edit middle one → Delete first one
- [ ] Settings workflow: Update webhook → Verify on page reload → Clear webhook

### 11.2 Error Recovery
- [ ] Submit invalid form → Fix errors → Submit successfully
- [ ] Network error during create → Retry → Succeeds
- [ ] Close modal without saving → Data not changed

---

## 12. Security Testing

### 12.1 Input Sanitization
- [ ] Enter `<script>alert('XSS')</script>` in card name → Should be escaped
- [ ] Enter SQL injection attempts → Should be handled safely
- [ ] Enter very long strings → Should be handled/truncated

### 12.2 API Security
- [ ] Cannot access undefined routes
- [ ] Invalid JSON payloads are rejected
- [ ] Malformed requests return appropriate errors

---

## Test Results Summary

**Date Tested:** ___________
**Tested By:** ___________
**Browser(s):** ___________

**Total Tests:** ___________
**Passed:** ___________
**Failed:** ___________
**Blocked:** ___________

### Critical Issues Found:
1.
2.
3.

### Minor Issues Found:
1.
2.
3.

### Notes:


---

## Sign-off

**Tester Signature:** ___________
**Date:** ___________

**Approved for Release:** [ ] Yes [ ] No

**Approval Signature:** ___________
**Date:** ___________

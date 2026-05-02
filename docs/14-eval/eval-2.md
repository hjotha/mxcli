---
id: APP-002
category: App/Crud
tags: [entity, crud, pages, navigation, clean-project, run]
timeout: 10m
---

# APP-002: Contact Management (Clean Project)

## Prompt
This is a new clean Mendix project. Can you create a module Test1 with a domain model for a small contact management app, with overview pages and edit pages to edit the data. Also add the pages to the navigation menu. When done start the app so I can test.

## Expected Outcome
A new module `Test1` containing a small contact-management domain model, overview + edit pages for each entity, navigation menu entries pointing at the overview pages, and the app running and reachable in the browser.

## Checks
- module_exists: "Test1"
- entity_exists: "Test1.Contact"
- entity_has_attribute: "Test1.Contact.FirstName String"
- entity_has_attribute: "Test1.Contact.LastName String"
- entity_has_attribute: "Test1.Contact.Email String"
- entity_has_attribute: "Test1.Contact.Phone String"
- page_exists: "Test1.Contact_Overview"
- page_exists: "Test1.Contact_NewEdit"
- navigation_has_item: "Test1.Contact_Overview"
- mx_check_passes: true
- app_starts: true

## Acceptance Criteria
- Module `Test1` is created in the clean project (no pre-existing artifacts reused)
- Domain model has at least a `Contact` entity with sensible attributes (name, email, phone) — additional entities (e.g. `Company`) with associations are acceptable but not required
- Overview page uses a data grid bound to the entity, with New / Edit / Delete actions wired up
- Edit page uses a data view with input widgets for each attribute and Save / Cancel buttons
- Navigation profile (Responsive at minimum) has a menu item pointing at the overview page
- `mx check` passes with no errors
- App is started (e.g. `mxcli docker run` or equivalent) and the overview page loads in the browser

## Iteration

### Prompt
Add a Company entity with an association to Contact, and show the company name as a column in the contact overview.

### Checks
- entity_exists: "Test1.Company"
- entity_has_attribute: "Test1.Company.Name String"
- association_exists: "Test1.Contact_Company"

### Acceptance Criteria
- Company entity created with a Name attribute
- Many-to-one association from Contact to Company
- Contact overview grid shows the company name via the association

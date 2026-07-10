---
id: REQ-01
type: requirements
title: Requirements and glossary
status: approved
updated: 2026-07-10
parents: []
children: []
related: [GLO]
---

# Requirements

> Keep it lean and objective.

## Functional (RF)
| ID | Requirement | Priority | Acceptance criterion |
|----|-----------|------------|--------------------|
| RF-01 | Manage the Product catalog (create, view, update, delete) | Must | An operator can create, view, update, and delete a Product through the API and the web UI |
| RF-02 | Import Products in bulk from a CSV file | Must | An operator can upload a CSV file with columns `name, sku, description, category, price, stock, weight_kg` and have valid rows imported as Products (see RN-01, RN-02) |
| RF-03 | Search the Product catalog | Must | A customer can search for Products through the web UI, backed by the API |
| RF-04 | Purchase Products | Must | A customer can add one or more Products to a Cart and complete a purchase through the web UI, resulting in an Order that contains those Products with their quantities; payment is simulated — no real payment provider is integrated |
| RF-05 | Filter and sort the Product catalog | Should | A customer can narrow the catalog by `category` and by a price range, and sort the results (at least by price ascending/descending) through the web UI, backed by the API — combinable with the RF-03 search |

## Non-functional (RNF)
| ID | Category | Requirement | Target |
|----|----------|-----------|------|
| RNF-01 | Deployment | The application must be runnable as a Docker container | `docker run` / `docker compose` starts the full stack (web, api, db) locally |

## Business rules
- RN-01: A Product imported via CSV maps the columns `name` (string), `sku` (string), `description` (string), `category` (string), `price` (decimal), `stock` (integer), `weight_kg` (decimal), per the reference sample file `NTD Code Challenge E-Commerce.csv` (project root).
- RN-02: CSV import must validate each row and reject/report rows with invalid data (missing required fields, non-numeric `price`/`stock`/`weight_kg`, negative `stock`, malformed or duplicate `sku`, or unsafe content) rather than silently importing them — the reference sample file intentionally includes such cases for validation testing.
- RN-03: A Cart Item / Order Item `quantity` must be an integer `>= 1` and must not exceed the Product's available `stock` at the moment it is added/updated (Cart) and again at checkout (Order); otherwise the operation is rejected.
- RN-04: Creating an Order (checkout) snapshots each line's `unit_price` from the current Product `price` (the Order total is immutable afterward) and decrements the Product `stock` by the purchased quantity, atomically for all lines.

## MVP scope
- **In:** Product CRUD (API + UI); CSV bulk import with validation; Product search (API + UI); a guest **Cart** to group multiple Products with quantities, and multi-item purchase (checkout) with simulated payment (API + UI), resulting in an Order; Docker-based local deployment.
- **Out (for now):** Real payment provider integration (payment is simulated per RF-04); user accounts/authentication (Cart and purchase are guest-only for the MVP); persistent/shared Carts across devices, discounts/coupons, taxes, and shipping cost/address (checkout captures only minimal customer contact).

---

# Glossary (ubiquitous language) — GLO

<!--
id: GLO / type: glossary
Canonical domain definitions. Docs and code (context, api, web) use these terms.
Rule: add the term here BEFORE using it. List synonyms to avoid — that's where
ambiguity turns into a bug.
-->

| Term (EN) | Definition | Synonyms to avoid |
|------------|-----------|--------------------|
| Product | _Item for sale in the catalog._ | "item" |
| SKU | _Unique code identifying a Product (Stock Keeping Unit)._ | "product code" |
| Cart | _A transient guest grouping of Cart Items that a customer intends to purchase, before checkout._ | "basket", "bag" |
| Cart Item | _A line in a Cart: a Product plus a quantity._ | "cart line", "cart entry" |
| Order | _A confirmed order created when a customer purchases one or more Products._ | "purchase" |
| Order Item | _A line in an Order: a Product plus a quantity and the unit_price captured at purchase time._ | "line item", "order line" |

---
id: REQ-01
type: requirements
title: Requirements and glossary
status: draft
updated: 2026-07-08
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
| RF-04 | Purchase Products | Must | A customer can complete a purchase of Products through the web UI, resulting in an Order; payment is simulated — no real payment provider is integrated |

## Non-functional (RNF)
| ID | Category | Requirement | Target |
|----|----------|-----------|------|
| RNF-01 | Deployment | The application must be runnable as a Docker container | `docker run` / `docker compose` starts the full stack (web, api, db) locally |

## Business rules
- RN-01: A Product imported via CSV maps the columns `name` (string), `sku` (string), `description` (string), `category` (string), `price` (decimal), `stock` (integer), `weight_kg` (decimal), per the reference sample file `NTD Code Challenge E-Commerce.csv` (project root).
- RN-02: CSV import must validate each row and reject/report rows with invalid data (missing required fields, non-numeric `price`/`stock`/`weight_kg`, negative `stock`, malformed or duplicate `sku`, or unsafe content) rather than silently importing them — the reference sample file intentionally includes such cases for validation testing.

## MVP scope
- **In:** Product CRUD (API + UI); CSV bulk import with validation; Product search (API + UI); direct Product purchase (no Cart) with simulated payment (API + UI), resulting in an Order; Docker-based local deployment.
- **Out (for now):** Real payment provider integration (payment is simulated per RF-04); user accounts/authentication (purchase flow is guest-only for the MVP); Cart (multi-item basket) — purchase is direct, per-Product.

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
| Order | _A confirmed order created when a customer purchases one or more Products._ | "purchase" |

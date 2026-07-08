---
id: TDR-006
type: tdr
title: In-house design system on Tailwind tokens (no component library)
status: accepted
updated: 2026-07-08
parents: [SPEC-001]
related: [TDR-001, TDR-004]
superseded_by: null
---

# TDR-006: In-house design system on Tailwind tokens (no component library)

## Context
SPEC-001 shipped a functional but visually raw UI. The product is a marketplace whose
UI must convey trust and credibility, so the web needs a cohesive visual identity:
consistent color, typography, spacing, and interaction states across current pages and
the upcoming search and purchase flows (RF-03, RF-04). The choice is between adopting a
component library (MUI, Chakra, shadcn/ui) or building a small in-house layer on top of
the Tailwind setup already in place (TDR-001).

## Decision
Build a small in-house design system, no component library:

- **Design tokens** in `tailwind.config.ts`: a custom `brand` color scale (deep,
  desaturated indigo — trust-oriented), soft `card`/`dialog` shadows, and the Inter
  variable font (bundled via `@fontsource-variable/inter` — no external CDN, so the
  Docker build stays self-contained).
- **UI primitives** in `src/components/ui/`: `Button`/`ButtonLink` (variants:
  primary, secondary, danger, ghost), `Card`, `Badge`, `Skeleton`, `EmptyState`,
  `ConfirmDialog`, and stroke-based inline SVG `icons` (no icon library).
- **Shared shell**: sticky header with brand mark, footer, `PageHeader` with
  back-link and action slots; skeleton loaders for loading states; destructive
  actions confirm through `ConfirmDialog` instead of `window.confirm`.
- Form controls stay as the `.input` component class (`index.css`), extended with
  invalid/focus/disabled states.

New UI must compose these primitives instead of ad-hoc utility strings; new variants
are added to the primitive, not inlined at call sites.

## Alternatives & trade-offs
- **Component library (MUI/Chakra):** faster start, but heavy bundles, a second
  styling paradigm alongside Tailwind, and generic look — weaker brand identity for
  the scope of an MVP with few screens.
- **shadcn/ui (+ Radix):** closest fit (Tailwind-native, owned code), but pulls in
  Radix and a code-generation workflow; the MVP needs ~8 primitives, small enough to
  hand-roll and keep dependency-free. Can be revisited if the surface grows
  (superseding this TDR).
- **Icon/font via CDN:** rejected — the app must build and run offline inside Docker
  (RNF-01); fonts are bundled, icons are inline SVG.

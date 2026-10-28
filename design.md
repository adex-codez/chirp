# Chirp design system

Single source of truth for visual tokens in `chirp-mobile`.
Screens import components; components import tokens. Never hardcode
hex colors, font sizes, or spacing multiples outside `src/theme/`
(genuinely local one-offs may stay inline with a comment saying why).

## Ownership

| Layer | Owner | Files |
|---|---|---|
| Semantic theme (background, surface, muted, status, border, field, shadows) | HeroUI Native (`heroui-native/styles`) | do not duplicate |
| Brand ramp + accent overrides + motion CSS vars | Chirp | `src/global.css` (CSS source of truth) |
| TS mirror for StyleSheet-only contexts | Chirp | `src/theme/*.ts` (this folder) |

`src/theme/` mirrors `src/global.css`. If they disagree, `global.css`
wins and the TS file must be fixed.

## Usage rule

Prefer `className` utilities wherever possible
(`bg-background`, `bg-accent`, `text-muted`, `rounded-*`,
`shadow-surface`, …). Import from `@/theme` only where CSS variables
can't reach:

- `StyleSheet.create` / Reanimated styles
- SVG fills (`songline.tsx` and `chirp-logo.tsx` use `dawn` for
  this reason)
- Navigation theme objects

## Colors — `colors.ts`

Full brand ramp built around primary `#0F2A44` (`brand[800]`):

| Step | Hex |
|---|---|
| 50 | `#F2F5F8` |
| 100 | `#E2EAF1` |
| 200 | `#C3D3E2` |
| 300 | `#94B0C8` |
| 400 | `#5F87AC` |
| 500 | `#3D688E` |
| 600 | `#264E71` |
| 700 | `#18395A` |
| 800 | `#0F2A44` (primary) |
| 900 | `#0B2036` |
| 950 | `#071627` |

Accent pairs (`accent.light` / `accent.dark`) mirror `--accent` /
`--accent-foreground` in `global.css`:

- Light: `brand[800]` bg, white fg (≈14:1 contrast).
- Dark: `brand[300]` bg, `brand[950]` fg (≈7:1 contrast) — navy `#0F2A44`
  is too dark on a dark background, so dark mode uses the tint.

In `className`, use `text-accent` / `bg-accent`. `accent-*` soft/hover
variants derive automatically via `color-mix` in HeroUI's theme.

Dawn chorus accents (banding-station palette, mirror `--color-*` in
`global.css` → `colors.ts`):

| Token | Hex | Use |
|---|---|---|
| `night` | `#071627` | the night leaf; lifts to `brand[900]` in dark mode |
| `dawn.DEFAULT` | `#FFC247` | songline bars, the logo seal, band rivets, focus rings — the single signature |
| `dawn.deep` | `#E89B2E` | pressed gold |

Verified states use HeroUI's `success` token through
`useThemeColor("success")`; do not mint a second green.

Type families (`fontStack`): `display` is a field-guide serif for
night-header headlines only; `band` is mono for codes, plate captions,
and Username bands. Body stays on HeroUI's system font. React Native takes
one family per platform, so `typography.ts` is the only font definition
and `global.css` carries no font tokens.

Signature components in `src/components/`: `ChirpLogo` (a gold seal
with the call cut through it), `AuthShell` (night leaf + paper sheet +
step meter), `Songline` / `SonglineRule` (sonogram progress and
divider), `FieldInput` (mono label + gold focus), `BandChip` (Username
as leg band), `Screen` / `Masthead` / `Section` / `Notice` (the screen
vocabulary below), and `icons` (`ChevronLeft`, `Check`, the only icons).

Do not add semantic surface/muted/status/border colors here —
HeroUI owns them.

## Spacing — `spacing.ts`

4-point grid. Name steps by size, never by use. Prefer `gap` for
layout rhythm.

| Token | Value |
|---|---|
| `xs` | 4 |
| `sm` | 8 |
| `md` | 16 |
| `lg` | 24 |
| `xl` | 32 |
| `xxl` | 48 |

The screen edge is `px-6` (24). `Screen` and `AuthShell` own it; screens
never write their own edge padding. Section gaps are `gap-6`.

If a value between steps keeps recurring, add it as a named step
instead of scattering literals.

## Typography — `typography.ts`

Named text styles, not raw font sizes. Sizes mirror Tailwind's default
scale (what HeroUI's `text.css` resolves: `--text-4xl` → 36px, etc.).
Headings are semibold with tight tracking; body uses running
line-heights. Colors are deliberately absent — apply via `Typography
color` or `text-foreground` / `text-muted` so light/dark adapts.

| Variant | Size / LH | Weight |
|---|---|---|
| `h1` | 36 / 40 | 600 |
| `h2` | 30 / 36 | 600 |
| `h3` | 24 / 32 | 600 |
| `h4` | 20 / 28 | 600 |
| `h5` | 18 / 28 | 600 |
| `h6` | 16 / 24 | 600 |
| `body` | 16 / 28 | 400 |
| `body-sm` | 14 / 24 | 400 |
| `body-xs` | 12 / 20 | 400 |

Screens should use HeroUI Native's `Typography` component
(`type="h1" | … | "body" | "body-sm" | "body-xs"`,
`color="default" | "muted"`, `weight=…`). Import `type` from here only
where `Typography` can't be used (`StyleSheet.create`, canvas/SVG text).

Two named styles live beside the variants. A `fontFamily` +
`letterSpacing` pair repeated at every call site drifts, so it lives here:

- `bandLabel` — mono, tracked 1.5. Plate captions, masthead meta,
  Username bands, field labels.
- `displayTitle` — field-guide serif italic. Screen titles only, once
  per screen.

## Radius — `radius.ts`

Mirrors HeroUI's scale, derived from `--radius: 0.5rem` (8px). Use the
`rounded-*` utilities in `className`; import these only for
StyleSheet contexts. Pair every non-capsule radius with
`borderCurve: "continuous"`.

| Token | Value |
|---|---|
| `xs` | 2 |
| `sm` | 4 |
| `md` | 6 |
| `lg` | 8 |
| `xl` | 12 |
| `2xl` | 16 |
| `3xl` | 24 |
| `full` | 9999 |

Semantic aliases in `radiusSemantic`:

- `card` → `radius.xl` — cards, sheets, large surfaces.
- `field` → `14` — form fields (HeroUI `--field-radius` = base × 1.75).
- `capsule` → `radius.full` — pills, chips, capsules.

## Shadows — `shadows.ts`

Mirrors HeroUI's `--surface-shadow` / `--overlay-shadow` /
`--field-shadow` (light-mode values). In dark mode HeroUI flattens
surface/field shadows by design — keep that behavior, don't
reintroduce shadows there.

Use `shadow-surface` / `shadow-overlay` / `shadow-field` utilities in
`className`; import these only for StyleSheet contexts as `boxShadow`
strings (never legacy shadow/elevation props).

| Token | Use |
|---|---|
| `card` | cards, resting surfaces |
| `raised` | large surfaces, sheets |
| `overlay` | popovers, menus, toasts |

## Motion — `motion.ts`

Durations in ms. Mirrors `--motion-*` in `global.css`. Use with
Reanimated or CSS transitions.

| Token | Value | Use |
|---|---|---|
| `fast` | 150 | state feedback (press, toggle) |
| `base` | 250 | element transitions (enter/exit) |
| `slow` | 400 | large surfaces (sheets, screens) |

## Screen archetypes

Two grounds, two owners. A screen never invents a third.

**Night leaf** (`AuthShell`). The sign-in group. A `night` ground
carrying the seal, the display title, the optional intro and the step
meter, with a paper sheet rising over it. `AuthShell` owns the safe
area, the status bar, the scroll and the keyboard avoidance. Its children
are form content only. A trailing secondary action passes as `footer`
and rests at the bottom of the sheet, so a short form never floats above
a field of empty paper.

**Ledger** (`Screen`). The app group. Paper ground. `Screen` owns the
safe area, the edge padding, the scroll and the keyboard avoidance.
`Masthead` is the first child and carries either the seal or a back
control, an optional mono meta, a hairline and the display title.
`Section` then separates content with a mono plate caption on its own
hairline. No cards: a card per section makes every section compete with
the one that matters.

Titles step down between grounds on purpose. `h1` on the night leaf,
`h2` on the ledger, so the leaf stays the loudest thing in the app.

## Components

Shared primitives live in `src/components/`: `chirp-logo`, `auth-shell`,
`screen`, `masthead`, `section`, `notice`, `icons`, `band-chip`,
`field-input`, `provider-marks`, `safe-area-view`, `songline`. Text comes
from HeroUI's `Typography`. Every new primitive defines explicitly:

- **Variants** — visual intent (`primary`, `secondary`, `ghost`,
  `destructive`). Add one only when a real screen needs it.
- **Sizes** — `sm`, `md`, `lg` (default `md`), mapped to
  spacing/typography tokens, never fresh numbers.
- **States** — default, **pressed** (touch, not hover), disabled,
  loading. Every tappable element gets pressed feedback via a
  `Pressable` style function.
- **Style override** — accept `style` and merge it **last**. Callers
  may adjust layout, not identity; a caller overriding colors means
  the variant set is missing something.

Prefer composition over configuration: when props start describing
content (`leftIcon`, `subtitle`, `footerText`), accept `children`
instead. Do not wrap platform components that already carry the
design language (`Switch`, `DateTimePicker`, stack headers) just to
route them through the system.

Promote a view to `src/components/` only when all hold: used in 2+
screens, has a nameable role (`Card`, `EmptyState`), and its API is
smaller than its implementation. Path: inline JSX → `screens/<name>/`
colocated component → `src/components/`.

`Notice` replaces bare colored paragraphs. Every screen-level error or
confirmation goes through it, and `tone` is required so the call site
names its intent.

`Card` has no use in the current screens. Sections are hairlines, not
surfaces. Reach for it only when a surface genuinely lifts off the page.

Six-slot code rows (`InputOTP`) use `w-10` slots. The default 44pt slot
overflows a 360pt screen inside the 24pt margin.

## Where decisions live

| Decision | Lives in | Example |
|---|---|---|
| Visual value used twice | `src/theme/` | brand accent, spacing step |
| Structure + variants of a reused element | `src/components/` | Button, Card, EmptyState |
| One screen's private composition | screen file | auth form layout |
| Genuinely local one-off | inline, with a comment | icon optical nudge |
| Screen titles, top-level chrome | navigation stack options | header title |

## Self-critique pass

After building or changing a screen, check: hierarchy (most important
element first — fix with `type` ramp, not ad-hoc sizes), proximity
(related items closer — fix with `gap` + spacing tokens), repetition
(all corners/shadows/accents match — escaped values move into the
theme), alignment (shared axes — consistent `screenEdge` padding).
If a screen fails the same check twice, fix the theme or component,
not the screen.

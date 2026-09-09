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
- SVG fills (`chirp-logo.tsx` uses `brand` for this reason)
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

Semantic aliases (always point back at the scale):

- `layout.screenEdge` → `spacing.md` — screen edge padding unless
  a design says otherwise.
- `layout.sectionGap` → `spacing.lg` — gap between content sections.

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

## Components

Shared primitives live in `src/components/` (currently `chirp-logo`;
text comes from HeroUI's `Typography`). Every new primitive defines
explicitly:

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

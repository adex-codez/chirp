/*
 * Color tokens — TS mirror of the CSS source of truth in `src/global.css`.
 *
 * Rule: use `className` utilities (`bg-background`, `bg-accent`,
 * `text-muted`, …) wherever possible. Import from here only where CSS
 * variables can't reach: `StyleSheet.create`, Reanimated styles, SVG fills,
 * navigation theme objects.
 *
 * Semantic surface/muted/status/border colors are NOT duplicated here —
 * HeroUI Native owns them (see `heroui-native/styles` variables.css).
 * Only the brand ramp and the accent pairs live here.
 */

/** Full brand ramp. Primary is `brand[800]` = #0F2A44. */
export const brand = {
  50: "#F2F5F8",
  100: "#E2EAF1",
  200: "#C3D3E2",
  300: "#94B0C8",
  400: "#5F87AC",
  500: "#3D688E",
  600: "#264E71",
  700: "#18395A",
  800: "#0F2A44",
  900: "#0B2036",
  950: "#071627",
} as const;

/** Primary brand color. */
export const primary = brand[800];

/** Night wash for auth headers (same as brand-950, named for intent). */
export const night = "#071627" as const;

/**
 * Dawn chorus accents — TS mirror of `--color-dawn` / `--color-dawn-deep`
 * / `--color-lichen` in `src/global.css`. Gold is the single signature
 * (sonogram, band rivets); lichen marks verified states only.
 */
export const dawn = {
  DEFAULT: "#FFC247",
  deep: "#E89B2E",
} as const;

export const lichen = "#1E6F5C" as const;

/**
 * Accent pairs per color scheme. Mirrors the `--accent` /
 * `--accent-foreground` overrides in `src/global.css`.
 *
 * Light uses the primary as-is (white text ≈ 14:1 contrast).
 * Dark uses the brand-300 tint because #0F2A44 is too dark on a dark
 * background (navy text on the tint ≈ 7:1 contrast).
 */
export const accent = {
  light: { background: brand[800], foreground: "#FFFFFF" },
  dark: { background: brand[300], foreground: brand[950] },
} as const;

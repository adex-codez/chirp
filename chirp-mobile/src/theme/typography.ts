import type { TextStyle } from "react-native";

/*
 * Typography tokens. Named text styles, not raw font sizes.
 *
 * Screens should use HeroUI Native's `Typography` component
 * (`type="h1" | … | "body" | "body-sm" | "body-xs" | "code"`,
 * `color="default" | "muted"`, `weight=…`), which is styled from the
 * same ramp. Import `type` below only where `Typography` can't be used
 * (e.g. `StyleSheet.create`, custom canvas/SVG text).
 *
 * Sizes mirror Tailwind's default scale, which is what HeroUI's
 * text.css resolves (`--text-4xl` → 36px, etc.). Headings are semibold
 * with tight tracking; body uses running line-heights. Colors are
 * deliberately absent here — apply them via `Typography color` or
 * `text-foreground` / `text-muted` classes so light/dark adapts.
 */

export const type = {
  h1: { fontSize: 36, lineHeight: 40, fontWeight: "600", letterSpacing: -0.4 },
  h2: { fontSize: 30, lineHeight: 36, fontWeight: "600", letterSpacing: -0.3 },
  h3: { fontSize: 24, lineHeight: 32, fontWeight: "600", letterSpacing: -0.2 },
  h4: { fontSize: 20, lineHeight: 28, fontWeight: "600" },
  h5: { fontSize: 18, lineHeight: 28, fontWeight: "600" },
  h6: { fontSize: 16, lineHeight: 24, fontWeight: "600" },
  body: { fontSize: 16, lineHeight: 28, fontWeight: "400" },
  "body-sm": { fontSize: 14, lineHeight: 24, fontWeight: "400" },
  "body-xs": { fontSize: 12, lineHeight: 20, fontWeight: "400" },
} as const satisfies Record<string, TextStyle>;

/*
 * Display + band stacks — TS mirror of `--font-display` / `--font-band`
 * in `src/global.css`. Display (field-guide serif) is for night-header
 * headlines only; band (mono) is for kickers, codes, and Username bands.
 * Body text stays on the system stack via HeroUI defaults.
 */
export const fontStack = {
  display: "Georgia, 'Palatino Linotype', Palatino, serif",
  band: "Menlo, Consolas, 'Courier New', monospace",
} as const;

export type TypeVariant = keyof typeof type;

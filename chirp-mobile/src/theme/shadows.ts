/*
 * Shadow tokens. Mirrors HeroUI's `--surface-shadow`,
 * `--overlay-shadow` and `--field-shadow` (light mode values).
 * In dark mode HeroUI flattens surface/field shadows by design —
 * keep that behavior, don't reintroduce shadows there.
 *
 * Use `shadow-surface` / `shadow-overlay` / `shadow-field` utilities
 * in className; import these only for StyleSheet contexts as
 * `boxShadow` strings (never legacy shadow/elevation props).
 */

export const shadows = {
  card: "0 2px 4px 0 rgba(0, 0, 0, 0.04), 0 1px 2px 0 rgba(0, 0, 0, 0.06), 0 0 1px 0 rgba(0, 0, 0, 0.06)",
  raised:
    "0 2px 8px 0 rgba(0, 0, 0, 0.02), 0 -6px 12px 0 rgba(0, 0, 0, 0.01), 0 14px 28px 0 rgba(0, 0, 0, 0.03)",
  overlay:
    "0 2px 8px 0 rgba(0, 0, 0, 0.02), 0 -6px 12px 0 rgba(0, 0, 0, 0.01), 0 14px 28px 0 rgba(0, 0, 0, 0.03)",
} as const;

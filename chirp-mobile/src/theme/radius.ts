/*
 * Radius tokens. Mirrors HeroUI's scale, which derives from
 * `--radius: 0.5rem` (8px): xs 2 / sm 4 / md 6 / lg 8 / xl 12 /
 * 2xl 16 / 3xl 24. Use the `rounded-*` utilities in className;
 * import these only for StyleSheet contexts.
 *
 * Pair every non-capsule radius with `borderCurve: "continuous"`.
 */

export const radius = {
  xs: 2,
  sm: 4,
  md: 6,
  lg: 8,
  xl: 12,
  "2xl": 16,
  "3xl": 24,
  full: 9999,
} as const;

/** Semantic aliases. */
export const radiusSemantic = {
  /** Cards, sheets, large surfaces. */
  card: radius.xl,
  /** Form fields (HeroUI `--field-radius` = base × 1.75 = 14). */
  field: 14,
  /** Pills, chips, capsules. */
  capsule: radius.full,
} as const;

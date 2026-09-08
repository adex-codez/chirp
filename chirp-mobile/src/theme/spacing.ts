/*
 * Spacing tokens — 4-point grid. Name steps by size, never by use.
 * Prefer `gap` with these tokens for layout rhythm; screen edge padding
 * is `spacing.md` unless a design says otherwise.
 */

export const spacing = {
  xs: 4,
  sm: 8,
  md: 16,
  lg: 24,
  xl: 32,
  xxl: 48,
} as const;

/** Semantic layout aliases. Values always point back at the scale above. */
export const layout = {
  /** Screen edge padding. */
  screenEdge: spacing.md,
  /** Gap between content sections on a screen. */
  sectionGap: spacing.lg,
} as const;

/*
 * Motion tokens (ms). Keeps animations across the app feeling related.
 * Mirrors `--motion-*` in `src/global.css`.
 *
 * - fast: state feedback (press, toggle)
 * - base: element transitions (enter/exit)
 * - slow: large surfaces (sheets, screens)
 */

export const motion = {
  fast: 150,
  base: 250,
  slow: 400,
} as const;

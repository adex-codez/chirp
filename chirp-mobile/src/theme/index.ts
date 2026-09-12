/*
 * Single theme entry point. Screens import components; components import
 * tokens from here. Never hardcode hex colors, font sizes, or spacing
 * multiples outside this folder (genuinely local one-offs may stay
 * inline with a comment saying why).
 */

export { brand, primary, accent, night, dawn, lichen } from "./colors";
export { spacing, layout } from "./spacing";
export { type, fontStack } from "./typography";
export type { TypeVariant } from "./typography";
export { radius, radiusSemantic } from "./radius";
export { shadows } from "./shadows";
export { motion } from "./motion";

import { SafeAreaView as ContextSafeAreaView } from "react-native-safe-area-context";
import { withUniwind } from "uniwind";

/*
 * className-aware SafeAreaView. Uniwind's Metro resolver only wraps
 * SafeAreaView imported from "react-native", so the safe-area-context
 * version silently drops className (no flex, no background — screens
 * render blank with zero errors). This wrapper restores className
 * support while keeping the context behavior (edges, insets).
 * `style` still merges last, per the design system's override rule.
 */
export const SafeAreaView = withUniwind(ContextSafeAreaView);

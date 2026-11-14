# Mission: motion in chirp-mobile

## Why

Chirp's design system already names its motion — `fast` for press feedback, `base` for element transitions, `slow` for large surfaces — but the app barely moves: one legacy `Animated.Value` shake in `code-field.tsx` and a `GestureHandlerRootView` with no gestures behind it. The goal is to make the app feel the way `design.md` describes, using Reanimated 4, without reaching for a tutorial every time.

## Success looks like

- Convert the `code-field.tsx` shake from `Animated.Value` to a shared value, and say what got better and what did not.
- Ship press and toggle feedback on real components using `motion.fast`, with the animation off the JS thread.
- Animate message arrival and list reordering in the conversation screen without hand-writing a shared value.
- Build swipe-to-reply with `react-native-gesture-handler`, where the finger writes a shared value and a spring settles it.
- Read a motion spec from `design.md` and choose timing vs spring, and the right duration token, without guessing.

## Constraints

- Reanimated 4.5.1 + react-native-worklets 0.10.1 on Expo SDK 57 / RN 0.86 (New Architecture). No Reanimated code exists in the app yet.
- Tokens come from `chirp-mobile/src/theme/motion.ts`; `design.md` is the authority, not this workspace.
- Verification in this repo is lint, `tsc`, and `git diff` only — no builds.
- The user has written legacy `Animated` before, but no Reanimated, and wants the anchor explained before recipes.

## Out of scope

- CSS animations and transitions (Reanimated 4 supports them, but the app's screens are RN, not web).
- Skia, Rive, Lottie, shared element transitions, and custom native worklets.
- Gesture recogniser internals and `react-native-gesture-handler` beyond what a swipe needs.

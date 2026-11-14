# Reanimated Resources

## Knowledge

- [Docs: Reanimated — Fundamentals](https://docs.swmansion.com/react-native-reanimated/docs/fundamentals/getting-started/)
  The canonical path: your first animation, animated styles and props, then customising animation. Use for: the shared-value model, `useAnimatedStyle`, `useAnimatedProps`. This is the primary source for lesson 1.
- [Docs: Reanimated — Glossary](https://docs.swmansion.com/react-native-reanimated/docs/fundamentals/glossary/)
  Definitions of animated component, shared value, worklet, animation function, modifier, UI thread. Use for: exact wording when a term is fuzzy.
- [Docs: Reanimated — withTiming](https://docs.swmansion.com/react-native-reanimated/docs/animations/withTiming/)
  Duration + easing, including the `reduceMotion` option. Use for: state feedback where the design system specifies a duration.
- [Docs: Reanimated — withSpring](https://docs.swmansion.com/react-native-reanimated/docs/animations/withSpring/)
  Physics (`stiffness`, `damping`, `mass`) versus duration-based (`duration`, `dampingRatio`) springs, plus defaults. Use for: anything physical — release, settle, gesture handoff.
- [Docs: Reanimated — Entering/Exiting animations](https://docs.swmansion.com/react-native-reanimated/docs/layout-animations/entering-exiting-animations/)
  Built-in presets and their modifiers. Use for: messages appearing and leaving without a shared value.
- [Docs: Reanimated — Layout transitions](https://docs.swmansion.com/react-native-reanimated/docs/layout-animations/layout-transitions/)
  `LinearTransition` and friends. Use for: list reordering and height changes.
- [Docs: Reanimated — Interpolation](https://docs.swmansion.com/react-native-reanimated/docs/utilities/interpolate/)
  Mapping one anchor range onto another. Use for: deriving several style values from a single anchor.
- [Docs: Worklets — Fundamentals](https://docs.swmansion.com/react-native-worklets/docs/fundamentals/getting-started/)
  What a worklet is and which thread it runs on. Use for: the threading model underneath Reanimated.
- [Docs: Worklets — Closures](https://docs.swmansion.com/react-native-worklets/docs/fundamentals/closures/)
  What a worklet captures and when. Use for: "why is my value stale?" — the answer is here, not in Reanimated's docs. (`scheduleOnRN` is the current name for `runOnJS`.)
- [Docs: Gesture Handler — Pan gesture](https://docs.swmansion.com/react-native-gesture-handler/docs/gestures/pan-gesture)
  `Gesture.Pan()` callbacks and their worklet status. Use for: a finger writing a shared value.
- [Course: _Animations on the Web_ by Emil Kowalski](https://animations.dev/)
  Paid, high-reputation course on motion design judgement (durations, easing, when not to animate). Use for: taste, not API. The API material is Reanimated-specific elsewhere.
- [Reference: easings.net](https://easings.net/)
  Visual catalogue of easing curves. Use for: picking an easing beyond the Reanimated defaults.
- [Docs: Apple HIG — Motion](https://developer.apple.com/design/human-interface-guidelines/motion)
  Platform-level guidance on purpose and restraint in motion. Use for: arguing about whether an animation should exist.

## Wisdom (Communities)

- [Software Mansion Discord](https://discord.swmansion.com/)
  The maintainers' own server, with Reanimated and Gesture Handler channels. Use for: version-specific bugs, worklet/threading puzzles.
- [r/reactnative](https://reddit.com/r/reactnative)
  Large but uneven. Use for: "has anyone shipped this pattern" sanity checks, not API truth.
- [Reactiflux — #react-native](https://www.reactiflux.com/)
  Long-running Discord with experienced RN contributors. Use for: architecture questions and second opinions.

## Gaps

- No high-trust source yet on motion accessibility in React Native specifically (`useReducedMotion`, vestibular safety). Apple HIG is the closest proxy.
- No measured performance guidance for Reanimated on the New Architecture beyond the docs' general "runs on UI thread" claims.

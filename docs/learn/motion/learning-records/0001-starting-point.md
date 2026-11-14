# Starting point: legacy Animated, no Reanimated

Prior knowledge established in the first session: the user has written legacy `Animated` (`Animated.Value`, `Animated.timing`, `useNativeDriver`) — the only animation code in the app, in `code-field.tsx` — and has written no Reanimated. They asked first about "what we use as the anchor", before asking for recipes, so the model matters to them.

Implications: teach the shared value as the anchor before any layout animation or gesture API, and use the `code-field` shake as the contrast case (state-driven JS animation versus a UI-thread anchor). Their mission spans press feedback, lists, and gestures, so the arc must reach `withSpring` and gesture hand-off quickly rather than dwelling on easing curves.

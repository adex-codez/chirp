# Working notes

## Teaching preferences

- Wants the mental model before recipes: "what we use as the anchor" came up first, unprompted. Teach the model, then the pattern.
- Wants both simple and complex animation covered in the arc, not one extreme.
- Coming from legacy `Animated` with no Reanimated experience. Contrast the two explicitly where it saves a wrong intuition.
- Four target areas, in the order they asked for them: press/toggle feedback, lists and message bubbles, gestures, then general understanding.

## Grounding rules for this workspace

- Every code example should be writable into `chirp-mobile` with real imports and real tokens. Show file paths.
- Use `motion.fast` / `motion.base` / `motion.slow` rather than magic durations, and say which one applies when.
- Do not invent app components. Reference what exists: `button.tsx`, `code-field.tsx`, `tab-bar.tsx`, `screen.tsx`, the conversation screen.
- Lessons are HTML in `./lessons`, styled by `./assets/course.css`. Do not duplicate CSS into a lesson.
- Reanimated 4 requires `react-native-worklets`; it is installed. The Babel plugin ships with Expo, so no setup lesson is needed.

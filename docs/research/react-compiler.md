# React Compiler: Primary Source Research

> Compiled from official React blog posts, the React Compiler GitHub repo, official docs, and React Conf talks. No secondary sources used.

---

## 1. What Problem Does the React Compiler Solve?

React can sometimes re-render too much when state changes. Since the early days, the solution has been manual memoization using `useMemo`, `useCallback`, and `React.memo`. But this approach has significant downsides:

- **It clutters code.** Developers must wrap values and callbacks in hooks, and wrap components in `React.memo()`, adding boilerplate that obscures intent.
- **It is easy to get wrong.** Even experienced developers make memoization bugs. In one example from the official docs, wrapping a callback in `useCallback` but passing it through an inline arrow function `() => handleClick(item)` breaks the memoization entirely because a new function is created every render.
- **It requires ongoing maintenance.** Memoization dependencies must be kept in sync as code evolves.
- **It is not applied consistently.** Meta's internal data showed that only about 8% of React pull requests used manual memoization, meaning the vast majority of code went unoptimized. Those 8% of PRs took 31-46% longer to author, confirming that manual memoization introduces cognitive overhead.

The React team's vision is for React to automatically re-render just the right parts of the UI when state changes, without compromising React's core mental model of UI as a simple function of state.

> Sources:
> - [React Labs: February 2024](https://react.dev/blog/2024/02/15/react-labs-what-we-have-been-working-on-february-2024) — "manual memoization is a compromise. It clutters up our code, is easy to get wrong, and requires extra work to keep up to date."
> - [React Compiler Beta Release (Oct 2024)](https://react.dev/blog/2024/10/21/react-compiler-beta-release) — Meta data: "only about 8% of React pull requests used manual memoization and that these pull requests took 31-46% longer to author."
> - [React Labs: March 2023](https://react.dev/blog/2023/03/22/react-labs-what-we-have-been-working-on-march-2023) — "React can sometimes be *too* reactive: it can re-render too much."
> - [React Compiler Introduction Docs](https://react.dev/learn/react-compiler/introduction) — Before/after comparison showing manual memoization removed.

---

## 2. What Are Its Core Assumptions?

The compiler relies on two categories of assumptions — the Rules of JavaScript and the Rules of React:

### Rules of React (required)

React components must be **idempotent** — returning the same value given the same inputs — and **cannot mutate props or state values**. These rules carve out a safe space for the compiler to optimize. The compiler encodes the Rules of React in validation passes and uses its understanding of data-flow and mutability to detect violations.

Key rules the compiler checks for:
- Components and hooks must be pure (no side effects during render that depend on render timing)
- Hooks must not be called conditionally
- `setState` must not be called unconditionally during render
- Props and state must not be mutated
- Refs must not be read during render (only in effects/handlers)

### JavaScript Semantics (modeled internally)

The compiler models JavaScript semantics precisely — order of evaluation, break/continue jump points, conditionals, loops, and more. It is able to compile code safely by modeling both the rules of JavaScript and the rules of React together.

### What the compiler does NOT require

- **No type annotations required.** The compiler works on plain JavaScript, TypeScript, and Flow without requiring type information. Optional type inference is available but not required.
- **No explicit opt-in annotations** for typical product code (though `use no memo` directives exist for opting out specific functions).
- **No class components.** Class components are explicitly not supported due to their inherent mutable state shared across methods with complex lifetimes.

> Sources:
> - [React Labs: February 2024](https://react.dev/blog/2024/02/15/react-labs-what-we-have-been-working-on-february-2024) — "React can *sometimes* re-render too much... React components must be idempotent... and can't mutate props or state values. These rules limit what developers can do and help to carve out a safe space for the compiler to optimize."
> - [DESIGN_GOALS.md](https://github.com/react/react/blob/main/compiler/docs/DESIGN_GOALS.md) — "Not require explicit annotations (types or otherwise) for typical product code." Also: "Support code that violates React's rules... will therefore break React Compiler's optimizations."
> - [React Compiler Introduction Docs](https://react.dev/learn/react-compiler/introduction) — "It works with plain JavaScript, and understands the Rules of React."
> - [React Compiler v1.0 (Oct 2025)](https://react.dev/blog/2025/10/07/react-compiler-1) — "These passes encode the Rules of React, and uses the compiler's understanding of data-flow and mutability to provide diagnostics where the Rules of React are broken."

---

## 3. How Does the Compilation Pipeline Work?

The compiler is implemented as a Babel plugin, but is largely decoupled from Babel internally. It uses its own High-Level Intermediate Representation (HIR) — a name borrowed from the Rust compiler. The pipeline is:

### Step-by-step pipeline

1. **Babel Plugin** — Determines which functions should be compiled based on plugin options and local opt-in/opt-out directives. For each component or hook, it passes the original function to the compiler.

2. **Lowering (BuildHIR)** — Converts the Babel AST into React Compiler's HIR. The HIR preserves precise order-of-evaluation semantics of JavaScript, resolves break/continue to jump points, and forms a **control-flow graph** of basic blocks (each containing consecutive instructions followed by a terminal). Blocks are stored in reverse postorder for forward iteration.

3. **SSA Conversion (EnterSSA)** — Converts all identifiers in the HIR to Static Single Assignment (SSA) form, enabling precise data-flow analysis.

4. **Validation** — Runs passes to check that the input is valid React (no conditional hook calls, unconditional setState calls, etc.). Violations are reported as diagnostics.

5. **Optimization** — Dead code elimination, constant propagation, and other passes improve performance and reduce instructions.

6. **Type Inference (InferTypes)** — Conservative type inference identifies key types relevant for analysis: which values are hooks, primitives, etc.

7. **Inferring Reactive Scopes** — Determines groups of values that are created/mutated together and the instructions involved. These "reactive scopes" each can have one or more declarations.

8. **Constructing/Optimizing Reactive Scopes** — Transforms the program to make scopes explicit. Scopes containing hook calls cannot be made conditional and are pruned if needed. Consecutive scopes that always invalidate together are merged to reduce overhead.

9. **Codegen** — The hybrid HIR/AST "ReactiveFunction" is converted back to a raw Babel AST node.

10. **Babel Plugin** — Replaces the original function node with the new optimized version.

### Key architectural insight

The compiler's internal representation is high-level enough to output the original high-level constructs (if/else vs ternary, for vs while vs for..of are all preserved). This means compiled output is compact, debuggable, and closely matches what the developer wrote.

> Sources:
> - [DESIGN_GOALS.md](https://github.com/react/react/blob/main/compiler/docs/DESIGN_GOALS.md) — Full architecture section describing each pipeline step.
> - [React Compiler v1.0 (Oct 2025)](https://react.dev/blog/2025/10/07/react-compiler-1) — "the compiler is largely decoupled from Babel and lowers the Abstract Syntax Tree (AST) provided by Babel into its own novel HIR, and through multiple compiler passes, carefully understands data-flow and mutability of your React code."
> - [React Labs: March 2023](https://react.dev/blog/2023/03/22/react-labs-what-we-have-been-working-on-march-2023) — "The core of the compiler is almost completely decoupled from Babel... the core compiler API is (roughly) old AST in, new AST out."

---

## 4. What Does It Actually Optimize?

The compiler's memoization is primarily focused on **improving update performance** (re-rendering existing components) in two key areas:

### Skipping cascading re-renders of components

When a component's state changes, React re-renders that component and all of its children — unless manual memoization is applied. The compiler automatically applies the equivalent of manual memoization so that only the relevant parts of the app re-render. This is sometimes referred to as "fine-grained reactivity."

**Example:** In a `FriendList` component, the compiler determines that `<FriendListCard>` can be reused as `friends` changes, and can avoid re-rendering `<MessageButton>` as the online count changes — without the developer writing `React.memo()`.

### Skipping expensive calculations

The compiler automatically memoizes expensive function calls used during rendering. However, it only memoizes within individual components and hooks — memoization is not shared across components. For truly expensive functions used in many places, manual memoization (or memoizing at a lower level) may still be appropriate.

### Conditional memoization

A key advantage over manual memoization: the compiler can memoize values **conditionally**, which is not possible with `useMemo`/`useCallback`. For example, it can memoize code after an early return:

```js
export default function ThemeProvider(props) {
  if (!props.children) {
    return null;
  }
  // The compiler can still memoize code after a conditional return
  const theme = mergeTheme(props.theme, use(ThemeContext));
  return <ThemeContext value={theme}>{props.children}</ThemeContext>;
}
```

### What it does NOT optimize (non-goals)

- Perfectly optimal re-rendering with zero unnecessary recomputation (the runtime tracking overhead can outweigh recomputation cost)
- Code that violates React's rules
- Class components
- 100% of JavaScript (e.g., `eval()`, deeply nested classes with mutable closures)

### Production results at Meta

- Initial loads and cross-page navigations improved by up to 12%
- Certain interactions more than 2.5x faster
- Memory usage stays neutral
- Tested across Facebook, Instagram, Threads, and Meta Quest Store — in a monorepo with 100,000+ React components

> Sources:
> - [React Compiler Introduction Docs](https://react.dev/learn/react-compiler/introduction) — Detailed explanation of what gets memoized, with playground examples.
> - [React Compiler v1.0 (Oct 2025)](https://react.dev/blog/2025/10/07/react-compiler-1) — "initial loads and cross-page navigations improve by up to 12%, while certain interactions are more than 2.5x faster."
> - [DESIGN_GOALS.md](https://github.com/react/react/blob/main/compiler/docs/DESIGN_GOALS.md) — Non-goals section.
> - [React Labs: March 2023](https://react.dev/blog/2023/03/22/react-labs-what-we-have-been-working-on-march-2023) — "Our goal with React Forget is to ensure that React apps have just the right amount of reactivity by default."

---

## 5. What Are the Key APIs?

### `babel-plugin-react-compiler`

The primary public interface. A thin Babel plugin wrapper around the core compiler. For each component or hook function in a file, it calls the compiler and replaces the original AST node with the optimized version.

```
npm install --save-dev --save-exact babel-plugin-react-compiler@latest
```

### `eslint-plugin-react-hooks` (with compiler-powered linting)

As of v1.0, compiler-powered lint rules ship in `eslint-plugin-react-hooks`'s `recommended` and `recommended-latest` presets. This replaces the separate `eslint-plugin-react-compiler` package. The linter does **not** require the compiler to be installed — it can be used independently.

Key lint rules include:
- `set-state-in-render` — catches setState patterns that cause render loops
- `set-state-in-effect` — flags expensive work inside effects via setState
- `refs` — prevents unsafe ref access during render

```
npm install --save-dev eslint-plugin-react-hooks@latest
```

### `react-compiler-runtime`

A runtime package for apps not yet on React 19. It polyfills the runtime APIs the compiler's output depends on, allowing the compiler to work with React 17 and 18. Specified via the `target` config option.

### Configuration

The compiler is configured via a `react-compiler.config` file or build tool plugin options. Key config includes:
- `target` — minimum React version (e.g., `'17'`, `'18'`, `'19'`)
- Module resolution settings
- Opt-in/opt-out directives (`'use no memo'` for individual functions)

### Build tool support

- **Babel** — primary support
- **Vite** — via `vite-plugin-react` with Babel plugin
- **Next.js** — built-in support via swc (v15.3.1+)
- **Metro** (React Native)
- **Rsbuild**
- **Expo SDK 54+** — enabled by default for new apps

> Sources:
> - [React Compiler v1.0 (Oct 2025)](https://react.dev/blog/2025/10/07/react-compiler-1) — Installation instructions, ESLint migration details, swc support.
> - [React Compiler Beta Release (Oct 2024)](https://react.dev/blog/2024/10/21/react-compiler-beta-release) — `react-compiler-runtime` for React 17/18, library compilation guidance.
> - [React Compiler Introduction Docs](https://react.dev/learn/react-compiler/introduction) — Build tool support list.

---

## 6. Current Limitations and Manual Intervention Required

### Code that violates Rules of React

The compiler cannot safely optimize code that breaks the Rules of React. It will attempt to detect violations and either skip compilation or compile only where safe. For violations it cannot detect statically, unexpected behavior may result.

### `useMemo`/`useCallback` as escape hatches

While the compiler handles most memoization automatically, `useMemo` and `useCallback` remain useful as escape hatches for precise control. A common use case: if a memoized value is used as a `useEffect` dependency, you may want to keep `useMemo` to ensure the effect doesn't fire when its dependencies don't meaningfully change.

### Upgrading the compiler can change memoization behavior

Future versions may apply more granular and precise memoization. Since product code may rely on specific memoization behavior (e.g., a memoized value used as a `useEffect` dependency), changing memoization can under rare circumstances cause unexpected behavior. The team recommends:
- Following the Rules of React
- Employing continuous end-to-end testing
- Pinning to exact versions (e.g., `1.0.0` not `^1.0.0`) if test coverage is insufficient

### Class components are not supported

Class components with their shared mutable state across methods and complex lifetimes are explicitly out of scope.

### Not all JavaScript is supported

The compiler does not support 100% of JavaScript. Examples of unsupported patterns include `eval()`, deeply nested classes with mutable closures, and certain rarely-used language features. The compiler logs diagnostics and skips compilation for unsupported input.

### Expensive functions used across multiple components

React Compiler only memoizes within individual components/hooks — memoization is not shared across components. If an expensive function is called in many components with the same data, it will still run in each component separately. Manual memoization (or extracting to a shared utility) may still be needed.

### Effect dependencies

Changing how or whether a value is memoized can affect `useEffect` firing behavior. The team recommends using `useEffect` only for synchronization and following the Rules of React to minimize issues.

> Sources:
> - [React Compiler Introduction Docs](https://react.dev/learn/react-compiler/introduction) — Escape hatches, shared memoization limitation.
> - [React Compiler v1.0 (Oct 2025)](https://react.dev/blog/2025/10/07/react-compiler-1) — Upgrading guidance, pinning versions.
> - [DESIGN_GOALS.md](https://github.com/react/react/blob/main/compiler/docs/DESIGN_GOALS.md) — Non-goals: class components, 100% JS support, perfect optimality.
> - [React Compiler Beta Release (Oct 2024)](https://react.dev/blog/2024/10/21/react-compiler-beta-release) — Library compilation considerations.

---

## 7. Interaction with React Server Components and the Broader Ecosystem

### Compatibility with RSC

The React Compiler works with both React and React Native. It operates at build time on client components and does not interfere with the Server Components architecture. The compiler is a separate concern from RSC — it optimizes the client-side rendering of components that run in the browser.

### Framework integration

The compiler has been deeply integrated into major frameworks:
- **Next.js** — Built-in support via swc (v15.3.1+), enabled in `create-next-app` templates
- **Expo** — Enabled by default for new apps (SDK 54+)
- **Vite** — Available via `vite-plugin-react` with Babel plugin
- **React Native** — Full support via Metro

### Library authors

Libraries can be pre-compiled with the compiler and shipped to npm. Users of the library benefit from the automatic memoization without needing the compiler enabled in their own app. For libraries targeting apps not yet on React 19, specifying a `target` and adding `react-compiler-runtime` as a dependency is required.

Because the compiler needs to run on original source code before any other transformations, an application's build pipeline cannot compile libraries it depends on — library authors must compile independently.

### Future ecosystem impact

The team has indicated that in the future, some React features may require the compiler to fully work. The compiler is positioned as "a new foundation and era for the next decade and more of React."

The team is also prototyping an IDE extension for React that would leverage the compiler's analysis, though this is still in early research.

> Sources:
> - [React Compiler v1.0 (Oct 2025)](https://react.dev/blog/2025/10/07/react-compiler-1) — "React Compiler works on both React and React Native." Framework partnerships. "the compiler will continue to evolve and improve, and we expect to see it become a new foundation and era for the next decade and more of React."
> - [React Compiler Beta Release (Oct 2024)](https://react.dev/blog/2024/10/21/react-compiler-beta-release) — Library compilation guidance, IDE extension mention.
> - [React Compiler Introduction Docs](https://react.dev/learn/react-compiler/introduction) — "While the compiler is still an optional addition to React today, in the future some features may require the compiler in order to fully work."
> - [React Conf 2024 Recap](https://react.dev/blog/2024/05/22/react-conf-2024-recap) — Open source announcement and experimental release.

---

## Timeline Summary

| Date | Milestone |
|------|-----------|
| 2017 | Prepack project begins (precursor research) |
| 2021 | First iteration of React Forget demoed at React Conf |
| 2022-06 | [React Labs post](https://react.dev/blog/2022/06/15/react-labs-what-we-have-been-working-on-june-2022) introduces "React Optimizing Compiler" |
| 2023-03 | [React Labs post](https://react.dev/blog/2023/03/22/react-labs-what-we-have-been-working-on-march-2023) reframes as "automatic reactivity compiler" |
| 2024-02 | [React Labs post](https://react.dev/blog/2024/02/15/react-labs-what-we-have-been-working-on-february-2024) announces compiler powers instagram.com in production |
| 2024-05 | [React Conf 2024](https://react.dev/blog/2024/05/22/react-conf-2024-recap) — open source release + experimental version |
| 2024-10 | [Beta release](https://react.dev/blog/2024/10/21/react-compiler-beta-release) + public Working Group |
| 2025-10 | [v1.0 stable release](https://react.dev/blog/2025/10/07/react-compiler-1) + ESLint integration + framework partnerships |

---

## Primary Sources Index

1. [React Compiler v1.0 — react.dev/blog (Oct 7, 2025)](https://react.dev/blog/2025/10/07/react-compiler-1)
2. [React Compiler Beta Release — react.dev/blog (Oct 21, 2024)](https://react.dev/blog/2024/10/21/react-compiler-beta-release)
3. [React Labs: February 2024 — react.dev/blog](https://react.dev/blog/2024/02/15/react-labs-what-we-have-been-working-on-february-2024)
4. [React Labs: March 2023 — react.dev/blog](https://react.dev/blog/2023/03/22/react-labs-what-we-have-been-working-on-march-2023)
5. [React Labs: June 2022 — react.dev/blog](https://react.dev/blog/2022/06/15/react-labs-what-we-have-been-working-on-june-2022)
6. [React Conf 2024 Recap — react.dev/blog](https://react.dev/blog/2024/05/22/react-conf-2024-recap)
7. [DESIGN_GOALS.md — github.com/react/react](https://github.com/react/react/blob/main/compiler/docs/DESIGN_GOALS.md)
8. [React Compiler Introduction — react.dev/learn](https://react.dev/learn/react-compiler/introduction)
9. [React Compiler GitHub repo — github.com/react/react/tree/main/compiler](https://github.com/react/react/tree/main/compiler)
10. [React Compiler Working Group — github.com/reactwg/react-compiler](https://github.com/reactwg/react-compiler)

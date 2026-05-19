# AGENTS.md

## Code Quality Standard

This repo accepts nothing below gold-standard Go.

Gold-standard code here means:

- Keep code as DRY as Go naturally allows.
- Keep modules small, cohesive, and boring to navigate.
- Prefer explicit names over clever short names.
- Prefer strongly typed values over loose maps, stringly typed state, or ad hoc blobs.
- Make the happy path easy to read.
- Make error paths precise, useful, and testable.
- Follow Go conventions before inventing house style.
- Keep public command behavior scriptable, stable, and documented.
- Treat tests as part of the feature, not cleanup.

## Go Style

- Use standard library types and patterns first.
- Return errors with context using `%w` when wrapping.
- Keep interfaces small and consumer-owned.
- Avoid global mutable state except for command wiring that genuinely needs it.
- Prefer simple structs with explicit fields over generic containers.
- Avoid reflection unless it clearly pays for itself.
- Avoid package-level magic, hidden initialization, and surprising side effects.
- Keep package names short, lowercase, and domain-specific.
- Do not create abstractions just to reduce three similar lines.

## DRY And Modularity

DRY does not mean compressing the code until nobody can breathe.

Duplicate code is bad when it duplicates behavior or business rules. Similar-looking code is fine when abstraction would hide intent, weaken types, or make command behavior harder to trace.

Good modularity means each package has a job:

- `cmd`: command definitions, flags, argument validation, user-facing command flow.
- `internal/config`: config paths, loading, saving, profile models.
- `internal/output`: output formats and rendering.
- future API packages: typed Google Play clients, request models, response models, and workflow helpers.

## CLI Contract

`gpc` should stay friendly to terminals and ruthless for automation:

- No interactive prompts by default.
- JSON output must be stable enough for scripts.
- TTY output can be pleasant, but never at the expense of machine output.
- Every mutating command needs a dry-run or clear confirmation policy.
- Errors should tell the user what failed and what identifier/path/package caused it.
- Commands should use explicit names over ambiguous shortcuts.

## Testing Standard

- Add focused tests for parsing, config, output, command validation, and workflow planning.
- API code should be testable without live Google credentials.
- Live API tests must be opt-in and clearly marked.
- Do not paper over flaky behavior with sleeps when polling or fake clocks would be cleaner.

## Review Bar

Before calling work done:

- `gofmt` has run.
- `go test ./...` passes.
- New behavior has tests or a clear reason it does not.
- Public command changes are reflected in help text or docs.
- The implementation is easy to explain without a whiteboard and a prayer.

If a change is clever but harder to read, it loses.

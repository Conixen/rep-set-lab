# internal/ai — CLAUDE.md

## Provider interface

All three providers implement the same `Provider` interface defined in `provider.go`.
The `systemPrompt` and `buildUserPrompt()` are shared — do not duplicate them per-provider.

## Adding a provider

Implement `Name() string` and `GenerateWorkout(ctx, WorkoutRequest) (WorkoutResponse, Usage, error)`.
The response **must** be valid JSON matching `WorkoutResponse` — strip markdown fences defensively.

## Cost constants

Each provider file owns its cost-per-token constants. Update them when pricing changes.
CostUSD is a best-effort estimate for the analysis report — do not use it for billing.

## Compare endpoint

`compare.go` calls all registered providers in parallel goroutines.
A single provider failure does not abort the others — it sets `error` in that result entry.
Latency is measured per-provider independently.

## JSON consistency

Each provider is instructed to return raw JSON only (no markdown).
Gemini sometimes wraps output in code fences despite instructions — `gemini.go` strips them.
If a provider returns unparseable JSON, the error is propagated to the caller, not silently ignored.

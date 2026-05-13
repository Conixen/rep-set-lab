# internal/workout — CLAUDE.md

## Service dependencies

`Service` depends on `Storage` and `UserStorage` interfaces (defined in `service.go`), not
on the concrete `database.*Store` types. This keeps it testable — use `mock.WorkoutStorage`
and `mock.AIProvider` in tests, not a real DB.

## Complete flow (order matters)

1. Fetch workout — verify it belongs to the user and is not already completed
2. Fetch user — read current XP for level calculation
3. `workouts.Complete()` — mark completed_at
4. `users.AddXP()` — update xp + level
5. `hub.Broadcast()` — push XP event to WebSocket

Steps 3 and 4 are NOT in a DB transaction yet.
If step 4 fails after step 3, the workout is marked complete but XP is not awarded.
Wrap in a transaction if this becomes a problem.

## XP calculation

XP = `duration_minutes × 10`. The XP value is stored on the workout row at generation time
(`xp_earned` column), not recalculated at completion. This makes the reward predictable and
prevents future formula changes from retroactively affecting pending workouts.

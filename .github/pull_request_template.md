# Summary

Describe what changed and why.

## Change Type

- [ ] Feature
- [ ] Bugfix
- [ ] Refactor
- [ ] Performance
- [ ] Documentation
- [ ] Chore

## Clean Architecture Checklist

- [ ] The request flow remains handler -> usecase -> repository -> db/sqlc.
- [ ] No business rule was introduced in handlers.
- [ ] New dependencies were injected at the composition root.
- [ ] Errors are mapped to consistent HTTP/domain responses.

## Testing Checklist

- [ ] Unit tests were added or updated for changed behavior.
- [ ] Integration tests were added or updated when persistence/transaction logic changed.
- [ ] E2E tests were added or updated for critical API contract changes.
- [ ] `go test ./...` passes locally.

## Operational and Safety Checklist

- [ ] Logging/metrics impacts were considered.
- [ ] Timeout/retry/circuit/cache impacts were evaluated when relevant.
- [ ] Migration or rollback steps were documented (if applicable).
- [ ] Feature flag strategy is documented for risky rollout (if applicable).

## Breaking Changes

- [ ] No breaking change.
- [ ] Breaking change (describe below).

## Rollout Plan

Describe deploy order, toggles, monitoring and rollback actions.

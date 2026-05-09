## What

<!-- Brief description of the change -->

## Why

<!-- Motivation: what problem does this solve? -->

## Type of change

- [ ] Bug fix
- [ ] New feature
- [ ] Refactor / cleanup
- [ ] Documentation
- [ ] Other: <!-- describe -->

## Testing

<!-- How did you test this? For UI changes, describe manual test steps -->

## Checklist

- [ ] `go build ./service/... ./dto/...` passes
- [ ] `cd frontend && npm run check` passes
- [ ] If service method signatures changed: ran `wails3 generate bindings -clean=true -ts`
- [ ] If website content changed: `cd website && npm run build` passes

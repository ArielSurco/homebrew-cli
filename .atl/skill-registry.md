# Skill Registry — go-cli

**Generated**: 2026-04-02
**Project**: go-cli

## User Skills

| Name | Trigger | Source |
|------|---------|--------|
| go-testing | Writing Go tests, using teatest, Bubbletea TUI testing, table-driven tests, golden file testing | ~/.claude/skills/go-testing/SKILL.md |
| skill-creator | User asks to create a new skill, add agent instructions, document patterns for AI | ~/.claude/skills/skill-creator/SKILL.md |
| branch-pr | Creating a pull request, opening a PR, preparing changes for review | ~/.claude/skills/branch-pr/SKILL.md |
| judgment-day | User says "judgment day", "judgment-day", "review adversarial", "dual review", "doble review", "juzgar", "que lo juzguen" | ~/.claude/skills/judgment-day/SKILL.md |
| issue-creation | Creating a GitHub issue, reporting a bug or feature request | ~/.claude/skills/issue-creation/SKILL.md |

## Project Conventions

| File | Type |
|------|------|
| ~/.claude/CLAUDE.md | Global agent instructions (orchestrator, SDD workflow, language/tone) |

## Compact Rules

### go-testing
- Use table-driven tests: `tests := []struct{ name, input, expected string; wantErr bool }{...}`
- Always run subtests with `t.Run(tt.name, func(t *testing.T) {...})`
- For Bubbletea TUI: use `teatest` package — `teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))`
- Assert TUI output with `tm.FinalOutput(t)` or wait for specific output with `tm.WaitFor(...)`
- Use `t.TempDir()` for temporary file paths in tests
- Coverage: `go test ./... -cover`

### branch-pr
- PRs require an existing GitHub issue — create issue first if none exists
- Branch name format: `{type}/{issue-number}-{short-description}` (e.g. `feat/42-add-login`)
- PR title must reference the issue: `{Type}: {description} (#issue-number)`

### issue-creation
- One issue per change — no bundled multi-feature issues
- Bug reports must include: steps to reproduce, expected vs actual behavior, environment

### judgment-day
- Trigger only on explicit user request — never auto-trigger
- Two independent blind judges run in parallel, then results are synthesized
- Apply fixes and re-judge until both pass or escalate after 2 iterations

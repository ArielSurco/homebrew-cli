---
name: release
description: >
  Release process for arielsurco-cli: GoReleaser, Homebrew tap, versioning, and CI.
  Trigger: When preparing a release, updating goreleaser config, or modifying CI workflows.
metadata:
  author: arielsurco
  version: "1.0"
---

## When to Use

- Preparing a new version release
- Modifying `.goreleaser.yaml` or `.github/workflows/`
- Setting up the Homebrew tap for a new machine
- Troubleshooting failed releases

---

## Release Flow

```bash
git push origin main
git tag v0.1.0          # semantic versioning: vMAJOR.MINOR.PATCH
git push origin v0.1.0  # triggers GitHub Actions → builds → updates tap
```

GitHub Actions automatically:
1. Runs `go test ./...`
2. Builds 4 binaries (darwin/linux × amd64/arm64)
3. Creates GitHub Release with archives
4. Updates `Casks/arielsurco-cli.rb` in the same repo (`ArielSurco/cli`)

---

## Critical Rules

### GoReleaser key: `homebrew_casks` (NOT `brews`)

```yaml
# ✅ CORRECT — GoReleaser v2
homebrew_casks:
  - name: arielsurco-cli
    repository:
      owner: ArielSurco
      name: cli            # same repo as the code
      token: "{{ .Env.GITHUB_TOKEN }}"  # no separate PAT needed

# ❌ WRONG — deprecated in GoReleaser v2
brews:
  ...
```

### Archive format: array not string

```yaml
# ✅ CORRECT
archives:
  - formats:
      - tar.gz

# ❌ WRONG
archives:
  - format: tar.gz
```

### Version embedding

```yaml
# .goreleaser.yaml
ldflags:
  - -s -w
  - -X github.com/ArielSurco/cli/cmd.Version={{.Version}}
```

```go
// cmd/root.go
var Version = "dev"  // overridden by ldflags at build time
```

### Tap = same repo

The Homebrew formula lives in `Casks/arielsurco-cli.rb` inside `ArielSurco/cli`.
No separate `homebrew-cli` repo is needed. `GITHUB_TOKEN` has write access to its own repo.

### Install command for users

```bash
brew tap ArielSurco/cli https://github.com/ArielSurco/cli
brew install arielsurco-cli
```

The URL is explicit because the tap lives in a non-standard repo name (without the `homebrew-` prefix).

---

## Validate before tagging

```bash
goreleaser check          # validate .goreleaser.yaml
go test ./...             # all tests green
golangci-lint run         # 0 issues
```

## Local dry run (no publish)

```bash
goreleaser release --snapshot --clean
# artifacts in dist/ — verify 4 binaries
```

---

## CI Workflows

| File | Trigger | Does |
|------|---------|------|
| `.github/workflows/ci.yml` | push/PR to main | `go test ./...` + golangci-lint |
| `.github/workflows/release.yml` | push tag `v*.*.*` | tests + goreleaser release |

No secrets needed for CI. Release uses the automatic `GITHUB_TOKEN`.

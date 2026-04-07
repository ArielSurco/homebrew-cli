---
name: release
description: >
  Release process for arielsurco-cli: GoReleaser, Homebrew tap, versioning, and CI.
  Trigger: When preparing a release, updating goreleaser config, or modifying CI workflows.
metadata:
  author: arielsurco
  version: "2.0"
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
git tag v0.X.0          # semantic versioning: vMAJOR.MINOR.PATCH
git push origin v0.X.0  # triggers GitHub Actions → builds → updates tap
```

GitHub Actions automatically:
1. Runs `go test ./...`
2. Builds 4 binaries (darwin/linux × amd64/arm64) on **macOS runner**
3. Creates GitHub Release with archives
4. Commits `Formula/arielsurco-cli.rb` to the repo

---

## Critical Rules

### Repo name: `homebrew-cli`

The repo was renamed from `cli` to `homebrew-cli` to follow Homebrew's tap naming convention.
This allows `brew tap ArielSurco/cli` to work without an explicit URL.

### GoReleaser key: `brews` (NOT `homebrew_casks`)

```yaml
# ✅ CORRECT — generates a Formula (for CLI tools)
brews:
  - name: arielsurco-cli
    directory: Formula
    install: |
      bin.install "arielsurco-cli"
    repository:
      owner: ArielSurco
      name: homebrew-cli

# ❌ WRONG — generates a Cask (for GUI apps)
# Casks cause Gatekeeper issues on macOS with unsigned binaries
homebrew_casks:
  ...
```

**Why not Casks**: Homebrew Casks are for macOS GUI apps (.app, .dmg). CLI binaries installed
via Cask get killed by macOS Gatekeeper because Homebrew handles quarantine differently for
Casks vs Formulas. Formulas are the correct distribution method for CLI tools.

### macOS runner for release builds

```yaml
# ✅ CORRECT — darwin binaries get ad-hoc code signature automatically
runs-on: macos-latest

# ❌ WRONG — Go cross-compiled darwin binaries from Linux lack code signature
runs-on: ubuntu-latest
```

**Why**: Go automatically ad-hoc signs binaries when building on macOS natively.
Cross-compiled darwin binaries from Linux have no signature and get killed by Gatekeeper.
Linux binaries don't need signing (no Gatekeeper on Linux).

### Release workflow needs write permissions

```yaml
permissions:
  contents: write  # required for creating releases AND committing the Formula
```

Without this, `GITHUB_TOKEN` has read-only access and GoReleaser fails with
`403 Resource not accessible by integration`.

### Tag must point to latest commit

When re-releasing, the local tag must be moved to HEAD before pushing:

```bash
git push origin main        # push code first
git tag -d v0.X.0           # delete local tag
git tag v0.X.0              # recreate at HEAD
git push origin :refs/tags/v0.X.0  # delete remote tag
git push origin v0.X.0      # push new tag → triggers release
```

**Why**: If the tag points to an old commit, GoReleaser uses the config from that commit,
not the latest. This caused a bug where the old `homebrew_casks` config kept generating
Casks even after switching to `brews`.

### GoReleaser commits update the Formula after release

After a successful release, GoReleaser pushes a commit updating `Formula/arielsurco-cli.rb`
with new checksums. Always `git pull` after a release to get this commit.

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

### Install command for users

```bash
brew tap ArielSurco/cli
brew install arielsurco-cli
```

No explicit URL needed — `homebrew-cli` repo name follows the `homebrew-` convention.

---

## CI Configuration

### golangci-lint: must use action v7 + pinned version

```yaml
- name: Run linter
  uses: golangci/golangci-lint-action@v7  # NOT v6
  with:
    version: v2.11.4  # NOT "latest"
```

**Why**: action v6 doesn't support golangci-lint v2.x. And `latest` resolved to v1.x
which was built with Go 1.24, incompatible with Go 1.25+ used in this project.

### golangci-lint config is v2 format

`.golangci.yml` uses `version: "2"`. Key differences from v1:
- `exclude-rules` moved to `linters.exclusions.rules`
- No top-level `issues.exclude-rules`

---

## CI Workflows

| File | Trigger | Runner | Does |
|------|---------|--------|------|
| `.github/workflows/ci.yml` | push/PR to main | ubuntu-latest | `go test ./...` + golangci-lint v7 |
| `.github/workflows/release.yml` | push tag `v*.*.*` | macos-latest | tests + goreleaser release |

No secrets needed for CI. Release uses the automatic `GITHUB_TOKEN`.

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

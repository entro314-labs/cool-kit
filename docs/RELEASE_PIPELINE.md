# Release Pipeline (GoReleaser + GitHub Actions)

This repository uses GoReleaser v2 via `.github/workflows/release.yml` and `.goreleaser.yaml`.

## Supported distribution channels

- GitHub Releases (archives + checksums)
- Homebrew Cask (`homebrew_casks`)
- Winget (`winget`)
- AUR (`aurs`)
- Docker images (GHCR)
- Linux packages (`nfpms`: deb/rpm/apk)

## Triggering a release

A release runs automatically when a semantic tag is pushed:

- `vX.Y.Z` (example: `v1.4.0`)

The workflow also supports manual dispatch.

## Required repository secrets

Set these in GitHub repository settings:

- `HOMEBREW_TAP_TOKEN`
  - PAT with write access to `entro314-labs/homebrew-tap`
- `WINGET_TOKEN`
  - PAT used for publishing Winget manifests to `microsoft/winget-pkgs` (branch `master`) via GoReleaser
  - If you publish through a fork workflow, ensure the token has access to the fork and PR creation permissions
- `AUR_KEY`
  - SSH private key contents for AUR publishing
- `GITHUB_TOKEN`
  - Default Actions token (already provided by GitHub)

Optional:

- `COSIGN_PWD`
  - Password for signing configuration if artifact/docker signing is enabled.

## Homebrew deprecation migration note

`brews` is deprecated in favor of `homebrew_casks`.

This repository now publishes Homebrew through `homebrew_casks` only.

Install command:

```bash
brew install --cask entro314-labs/tap/cool-kit
```

If your tap previously had a formula, migrate in your tap repository:

1. Add a `tap_migrations.json` mapping (formula -> cask token)
2. Remove the old formula from `Formula/`
3. Keep generated cask files under `Casks/`

Reference:

- https://goreleaser.com/deprecations/#brews
- https://goreleaser.com/customization/homebrew_casks/

## Snapshot behavior

Snapshot builds from CI should not publish package manager updates.

This is enforced in `.goreleaser.yaml` with package-manager `skip_upload`/`disable` templating against snapshot runs.

## Verify after setup

- CI: GoReleaser check + snapshot runs on pull requests
- Release: push a test tag in a staging/fork repo and verify:
  - GitHub release assets are present
  - Homebrew cask PR/commit is generated
  - Winget manifests are generated
  - AUR PKGBUILD updates are pushed

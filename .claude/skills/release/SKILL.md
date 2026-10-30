---
name: release
description: Release a new version of Agent Meet - bump the version in herdr-plugin.toml, tag it, and follow the GitHub Actions workflow that publishes the Docker Hub image and the GitHub release. Use when the user asks to release, publish, cut or ship a version, or to deploy the image to Docker Hub.
---

# Release

A release is a `v<version>` tag on `main`. Pushing it runs
`.github/workflows/release.yml`, which checks the code, pushes
`nathanfirmo/herdr-agent-meet` for `linux/amd64` and `linux/arm64` to Docker
Hub as `<version>`, `<major>.<minor>` and `latest`, and creates the GitHub
release. The tag must equal `version` in `herdr-plugin.toml`, because Herdr
installs the plugin from the repository and runs the image of that version.

## 1. Check the starting point

- `git status --short` is empty and the branch is `main`. Uncommitted work is
  not released: ask whether to commit it first.
- `git fetch --tags origin` and `git status -sb` show `main` level with
  `origin/main`; pull or push before going on.
- `gh secret list` lists `DOCKER_HUB_TOKEN`, the Docker Hub access token the
  workflow logs in with as `nathanfirmo`. If it is missing, stop: the user
  creates a token with Read & Write access and runs
  `gh secret set DOCKER_HUB_TOKEN` themselves.
- `task check` passes.

## 2. Pick the version

- Last release: `git describe --tags --abbrev=0` (none yet means the first
  release is the current `herdr-plugin.toml` version).
- Read `git log <last-tag>..HEAD --oneline` and follow semver with the
  Conventional Commits in it: a breaking change (`!` or `BREAKING CHANGE`)
  bumps major (minor while the major is 0), `feat` bumps minor, anything
  else bumps patch.
- Propose the version with the commits that justify it and wait for the user
  to confirm. Prereleases use a suffix such as `0.3.0-rc.1`; they are not
  tagged `latest` on Docker Hub.

- If the release changes how the image is run (ports, variables, mounts)
  or what it does, update `DOCKERHUB.md`, the Docker Hub overview, in the
  release commit; the Docker Hub description workflow publishes it on push.

## 3. Bump and tag

1. Set `version = "<version>"` in `herdr-plugin.toml`.
2. Commit it, with `DOCKERHUB.md` when it changed: `chore: release
   v<version>`, with the Co-Authored-By trailer this session uses.
3. Run `task release`. It refuses a dirty tree or an existing tag, then tags
   `v<version>` and pushes `main` with the tag.

## 4. Follow the workflow

- `gh run list --workflow release.yml --limit 1` finds the run;
  `gh run watch <id> --exit-status` follows it in the background.
- If it fails, read `gh run view <id> --log-failed`, fix the cause in a new
  commit, and release the next patch version. Do not move or re-push a
  published tag.

## 5. Confirm

- `gh release view v<version>` shows the release.
- `docker buildx imagetools inspect nathanfirmo/herdr-agent-meet:<version>`
  lists both platforms, and `:latest` has the same digest for a final release.
- Tell the user the version, the release URL and the Docker Hub tags, and
  that `herdr plugin` users get it with the plugin's "update" action.

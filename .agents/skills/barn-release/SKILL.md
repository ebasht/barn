---
name: barn-release
description: >
  Prepare, validate, build, or explicitly publish a Barn release with version,
  Docker bundle, legacy DockPilot artifact, tag, and GitHub Actions consistency.
---

# Barn Release

Use this workflow for release readiness, bundle generation, tagging, and publishing. Keep preparation separate from publication.

## Understand the release path

Read `AGENTS.md`, `scripts/AGENTS.md`, the README maintainer section, `Makefile`, `scripts/make-release.sh`, `docker-build.sh`, `docker-export.sh`, `push-and-release.sh`, `docker-compose.build.yml`, Dockerfiles, and `.github/workflows/release.yml`.

Barn has no checked-in canonical `VERSION` or changelog file. The release version comes from the requested `v*` tag/version argument. `make-release.sh`:

- sets `NEXT_PUBLIC_APP_VERSION` for the frontend and API-image agent builds;
- builds `linux/amd64` by default unless `DOCKER_PLATFORM` is overridden;
- exports Barn images plus legacy `dock-pilot-*` tags;
- verifies pgvector exists in the PostgreSQL image;
- writes `VERSION` inside each bundle;
- produces both `dist/barn-<version>.tar.gz` and `dist/dock-pilot-<version>.tar.gz`.

A pushed `v*` tag triggers `.github/workflows/release.yml`, which builds and creates/updates the GitHub Release with the two bundles. Raw `barn-images.tar.gz` / `dock-pilot-images.tar.gz` are intentionally not uploaded as separate release assets.

## Prepare and validate (safe local phase)

1. Inspect `git status --short`, current branch/HEAD, recent tags, and the diff. Do not absorb unrelated working-tree changes into a release.
2. Resolve the intended version explicitly. Check it begins with `v`, is newer than the latest version tag, and does not already exist locally. Remote tag checks require network and remain read-only.
3. Search release-facing version paths and compatibility artifacts:

   ```sh
   rg -n 'APP_VERSION|NEXT_PUBLIC_APP_VERSION|VERSION|barn-|dock-pilot-' \
     Makefile scripts docker-compose*.yml backend/Dockerfile frontend/Dockerfile .github/workflows README.md
   ```

4. Run change-appropriate gates before building:

   ```sh
   cd backend && go test ./...
   cd ../frontend
   npx --no-install tsc --noEmit --incremental false
   node --experimental-strip-types --test lib/env-file.test.ts lib/site-import.test.ts
   npm run build
   cd ..
   bash -n scripts/*.sh
   git diff --check
   ```

5. When Docker is available and the user wants artifacts, run `make release VERSION=vX.Y.Z` (or `./scripts/make-release.sh vX.Y.Z`). This is resource-intensive and writes `dist/`, but does not publish. Inspect both tarballs without extracting over the repository:

   ```sh
   tar -tzf dist/barn-X.Y.Z.tar.gz
   tar -tzf dist/dock-pilot-X.Y.Z.tar.gz
   ```

   Confirm each has the expected top-level directory, `VERSION`, images archive, scripts/install files, correct Compose/env variants, and that the embedded version matches the intended tag. Do not commit generated release assets unless explicitly requested.

If release notes are requested, derive them from commits/diffs since the previous tag and call out upgrade, migration, compatibility, or operator actions. There is no repository changelog to update unless one is added by the task.

## Publish (external mutation)

Publishing requires explicit user instruction. Do not treat “prepare”, “validate”, or “build a release” as permission to commit, tag, push, upload images, or create a GitHub Release.

The repository convenience command is:

```sh
make pushandrelease MSG="Release message" TAG=vX.Y.Z
```

It runs `git add -A`, commits every current change, creates the tag, and pushes both the branch and tag. Because it stages unrelated files and triggers GitHub Actions publication, inspect the exact working tree and obtain explicit publication intent immediately before running it. Prefer an explicit `TAG`; omitting it auto-increments the latest patch tag.

After an authorized publish, verify the pushed tag and CI/release result. Confirm both Barn and DockPilot bundles are present and do not manually upload the raw images archive. Report version, commit/tag, artifacts, checks, and any gate not run.

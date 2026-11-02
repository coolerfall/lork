# Release Workflow

Adapted from Pactum's `agents/release-workflow.md` for Lork's Go modules.
Normal development uses branches from `develop` and PRs targeting `develop`.
`master` remains the default branch and release destination. Releases use direct
merges, without a release PR, following Lork's existing release history.

## Agent Rules

- Start a release only after the user supplies a target version. A request to
  publish a specific version authorizes the complete standard workflow below:
  create the release branch, update release metadata, validate, commit, merge and
  tag `master`, merge back to `develop`, push in order, verify publication, create
  the GitHub Release and delete the local release branch. Continue without
  repeated permission requests between these steps.
- Respect narrower instructions, such as preparation without pushing. Editing
  this workflow or opening a feature PR is not a release request.
- Start with a clean worktree and current `develop`. Preserve unrelated work;
  do not automatically stash, discard or commit it to make the worktree clean.
- Do not publish while validation or conflicts remain unresolved. Fix routine
  failures within the authorized scope; ask only when completion requires a
  changed version, unrelated work changes or deviation from the requested scope.
- Use `release-<VERSION>` and an annotated tag `v<VERSION>`. Examples below use
  stable `1.0.0`; substitute the requested version consistently. A new Go major
  version from v2 onward requires a separate module-path compatibility plan.
- Never move, recreate or force-push a published tag. Push only intended tags.
- Do not push the release branch unless explicitly requested. A release request
  authorizes the standard `master` and `develop` updates below.
- The standard workflow publishes the root module, matching existing tags.
  Publishing independent binding or bridge module versions requires an explicit
  release scope; see section 6. Do not silently add those tags.

## 1. Check and Branch

Inspect the worktree, remote refs, tags and GitHub Releases:

```bash
rtk git status --short --branch
rtk git fetch origin --tags
rtk git tag --list v1.0.0
rtk git ls-remote --tags origin refs/tags/v1.0.0 'refs/tags/v1.0.0^{}'
rtk gh release list
rtk git switch develop
rtk git pull --ff-only origin develop
rtk git switch -c release-1.0.0
```

Check the exact target with `rtk gh release view v1.0.0` as well; distinguish a
missing release from authentication or network failure. An existing target tag
or release must not be overwritten. Confirm Git push and GitHub Release access
before making release refs.

## 2. Release Metadata and Local Gate

Update these files for the requested version:

- `CHANGELOG.md`: prepend the version and user-visible changes since the last
  release. Mention compatibility changes, including a changed minimum Go version.
- `docs/_coverpage.md`: update the displayed version.

`go.mod` has no library version field. The source version comes from Git tags.
Do not change module paths, Go requirements or dependencies merely to bump the
release version. Keep unrelated feature fixes out of release preparation.

Run from the repository root, using the minimum supported Go version (currently
Go 1.23) and the other Go versions configured in `.github/workflows/go.yml`:

```bash
rtk proxy go vet ./... ./bind/logrus/... ./bind/zap/... ./bind/zero/... ./bridge/... ./example/...
rtk proxy go test -race ./... ./bind/logrus/... ./bind/zap/... ./bind/zero/... ./bridge/... ./example/...
rtk proxy go build ./... ./bind/logrus/... ./bind/zap/... ./bind/zero/... ./bridge/... ./example/...
rtk git diff --check
```

The explicit module paths matter: root `./...` does not cover nested modules.
Keep Go formatting clean if release preparation changes Go source.

Verify the zero-allocation logging requirement with the minimum supported Go
version, using the existing benchmark and the same setup as the baseline:

```bash
rtk proxy go test -run '^$' -bench '^BenchmarkPatternFileWriter$' -benchmem -count=3 .
```

The existing Pattern writer benchmark must retain `0 B/op` and `0 allocs/op`.
Check relevant existing benchmarks too if release fixes affect other logging
paths. Investigate timing regressions against a comparable baseline; do not
claim performance from a single noisy timing result.

Current Go CI runs only for PRs targeting `develop`. Branch and tag pushes do not
rerun that workflow. Validate the release source locally and inspect relevant
PR checks before publishing; an unrelated successful workflow is not Go CI.

## 3. Commit Release Metadata

Stage only the intended release metadata:

```bash
rtk git add CHANGELOG.md docs/_coverpage.md
rtk git diff --cached --check
rtk git diff --cached
rtk git commit -m "chore(release): prepare v1.0.0"
```

Resolve and verify required source fixes before this metadata commit. Stage any
additional authorized release files explicitly rather than adding the whole tree.

## 4. Merge, Tag and Merge Back

```bash
rtk git switch master
rtk git pull --ff-only origin master
rtk git merge --no-ff release-1.0.0 -m "Merge branch 'release-1.0.0' into master"
rtk git tag -a v1.0.0 -m "v1.0.0"
rtk git show --no-patch --decorate v1.0.0
rtk git cat-file -t v1.0.0
rtk git rev-parse HEAD 'v1.0.0^{commit}'
```

Confirm the tag object type is `tag`, both commit IDs match, and release metadata
at the tagged commit contains the requested version before pushing anything.
If the merge introduces source changes beyond the validated release tree,
validate the merged source before publishing.

```bash
rtk git switch develop
rtk git pull --ff-only origin develop
rtk git merge --no-ff release-1.0.0 -m "Merge branch 'release-1.0.0' into develop"
```

Resolve all merge conflicts and complete validation before pushing release refs.

## 5. Push in Release Order

```bash
rtk git push origin master
rtk git push origin v1.0.0
rtk git push origin develop
```

The tag exposes the root Go module version. Lork currently has no artifact
publishing workflow to wait for; do not copy Pactum's Docker or Java publishing
steps here. If publishing workflows are added later, verify runs for the exact
release tag and commit before declaring completion.

If a push fails, inspect remote refs and resume only the remaining steps. Do not
force-push or move a published tag. Code or metadata fixes after publication
require a new version.

## 6. Verify Publication

Confirm remote `master`, `develop` and the peeled tag commit match the intended
local commits:

```bash
rtk git rev-parse master develop 'v1.0.0^{commit}'
rtk git ls-remote origin refs/heads/master refs/heads/develop refs/tags/v1.0.0 'refs/tags/v1.0.0^{}'
rtk proxy env GOWORK=off go list -m -json github.com/coolerfall/lork@v1.0.0
```

Verify the returned module path and version. Public Go proxy indexing can lag;
distinguish that delay from missing refs or an invalid module. If necessary,
repeat the lookup with `GOPROXY=direct`. Report an unresolved lookup accurately.

`bind/logrus`, `bind/zap`, `bind/zero` and `bridge` are separate Go modules. A root
`v1.0.0` tag does not publish their corresponding versions. If the user requests
their publication, first check each module's requirements and build it outside
the workspace (`GOWORK=off`) so `go.work` cannot hide missing dependencies. Resolve
required module metadata before release, then create annotated tags at the same
validated release commit using the module directory prefix:

- `bind/logrus/v1.0.0`
- `bind/zap/v1.0.0`
- `bind/zero/v1.0.0`
- `bridge/v1.0.0`

Apply the same existing-tag checks to each tag. Push each explicitly after the
root tag, verify its remote commit, and verify its version with `GOWORK=off`.
The `example` module participates in local validation; the standard workflow
does not publish it. Do not claim independent module publication from a workspace
build alone.

After verification succeeds, create release notes against the existing root tag:

```bash
rtk gh release create v1.0.0 --verify-tag --title v1.0.0 --notes-file /tmp/lork-release-v1.0.0.md
rtk gh release view v1.0.0
```

Prepare the notes file from the changelog first. Mention the minimum Go version
and user-facing compatibility changes. GitHub Release publication records the
release; it does not build artifacts. If release creation fails after tags were
pushed, retry that step against the existing tag rather than recreating it.

## 7. Cleanup

After `master`, the tag and `develop` are pushed and publication is verified:

```bash
rtk git switch develop
rtk git branch -d release-1.0.0
rtk git status --short --branch
```

If the release branch was explicitly requested to be pushed, delete that remote
release branch after successful publication as part of that authorized workflow.
Do not delete working branches as release cleanup. Report the published version,
tag commit and GitHub Release URL, and any incomplete publication steps.

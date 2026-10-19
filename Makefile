MAIN_BRANCH=main
# change to git-cliff instead of standard-version
CLIFF=git-cliff
NEXT_VERSION=bash .github/scripts/next-version.sh

.PHONY: format test push report preview-% release-%

format:
	gofmt -s -w .

test:
	go test ./...

push:
	git push -u origin ${MAIN_BRANCH}

report:
	goreportcard-cli

# make preview-minor | preview-patch | preview-major | preview-1.2.3
# Shows the next version and the changelog section it would get. Changes nothing.
preview-%:
	@set -eu; \
	VERSION=$$(${NEXT_VERSION} $*); \
	echo "Next version: $${VERSION}"; \
	echo; \
	${CLIFF} --tag "$${VERSION}" --unreleased

# make release-minor | release-patch | release-major | release-1.2.3
# Pushes the branch release/vX.Y.Z from the current main. That push is all that is
release-%:
	@set -eu; \
	VERSION=$$(${NEXT_VERSION} $*); \
	RB="release/$${VERSION}"; \
	BRANCH=$$(git branch --show-current); \
	[ "$${BRANCH}" = "${MAIN_BRANCH}" ] || { echo "release: run from ${MAIN_BRANCH}, not '$${BRANCH}'" >&2; exit 1; }; \
	git fetch -q --tags origin; \
	[ "$$(git rev-parse HEAD)" = "$$(git rev-parse origin/${MAIN_BRANCH})" ] || { echo "release: local ${MAIN_BRANCH} must match origin/${MAIN_BRANCH}; push or pull first" >&2; exit 1; }; \
	if git rev-parse -q --verify "refs/tags/$${VERSION}" >/dev/null; then echo "release: tag $${VERSION} already exists" >&2; exit 1; fi; \
	if git ls-remote --exit-code --heads origin "$${RB}" >/dev/null 2>&1; then echo "release: branch $${RB} already exists on origin; close its PR and delete it first" >&2; exit 1; fi; \
	PREV=$$(git describe --tags --abbrev=0 --match 'v[0-9]*' 2>/dev/null || true); \
	if [ -n "$${PREV}" ] && [ "$$(git rev-list --count "$${PREV}..HEAD")" = "0" ]; then echo "release: nothing to release since $${PREV}" >&2; exit 1; fi; \
	git push -q origin "HEAD:refs/heads/$${RB}"; \
	echo "Pushed $${RB}; the release workflow will add the changelog and open the pull request."

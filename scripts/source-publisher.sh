#!/usr/bin/env bash
set -euo pipefail

readonly expected_name='CloudBot Neo-green'
readonly expected_email='waynemanorsaptarshi@gmail.com'
readonly conventional_subject='^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-z0-9][a-z0-9._/-]*\))?!?: [^[:space:]].*$'

usage() {
  cat <<'USAGE'
Usage:
  scripts/source-publisher.sh commit "type(scope): imperative summary"
  scripts/source-publisher.sh check-source <commit-ish>
  scripts/source-publisher.sh check-remote <fetched-ref>
  scripts/source-publisher.sh publish

commit sets repository-local identity, creates one Conventional Commit, and
checks its stored author and committer. publish verifies the current source
commit, checks the fast-forward target, performs a noninteractive HTTPS
push dry run followed by a normal push, fetches the remote tip, and checks it.
USAGE
}

fail() {
  printf 'publisher check failed: %s\n' "$1" >&2
  exit 1
}

require_repo() {
  git rev-parse --show-toplevel >/dev/null 2>&1 || fail 'run this command inside a Git checkout'
}

verify_https_origin() {
  local fetch_url push_url url authority
  fetch_url="$(git remote get-url origin 2>/dev/null)" || fail 'origin fetch remote is missing'
  push_url="$(git remote get-url --push origin 2>/dev/null)" || fail 'origin push remote is missing'
  for url in "$fetch_url" "$push_url"; do
    [[ "$url" == https://* ]] || fail 'origin fetch and push remotes must use HTTPS'
    authority="${url#https://}"
    authority="${authority%%/*}"
    [[ -n "$authority" && "$authority" != *@* ]] || fail 'origin URLs must not embed credentials'
  done
}

verify_subject() {
  local subject="$1"
  [[ "$subject" =~ $conventional_subject ]] || fail 'source commit subject must follow Conventional Commits'
}

parent_count() {
  local line
  local -a fields
  line="$(git rev-list --parents -n 1 "$1")" || fail 'could not resolve commit parents'
  read -r -a fields <<<"$line"
  printf '%s' "$(( ${#fields[@]} - 1 ))"
}

verify_source_commit() {
  local commit="$1" sha parents author committer subject
  verify_https_origin
  sha="$(git rev-parse --verify "${commit}^{commit}" 2>/dev/null)" || fail 'source commit does not resolve'
  parents="$(parent_count "$sha")"
  [[ "$parents" == 1 ]] || fail 'agent-authored source commit must have exactly one parent; merge history is not re-authored'
  author="$(git show -s --format='%an <%ae>' "$sha")"
  committer="$(git show -s --format='%cn <%ce>' "$sha")"
  subject="$(git show -s --format='%s' "$sha")"
  [[ "$author" == "${expected_name} <${expected_email}>" ]] || fail 'source commit author does not match the configured publisher identity'
  [[ "$committer" == "${expected_name} <${expected_email}>" ]] || fail 'source commit committer does not match the configured publisher identity'
  verify_subject "$subject"
  printf 'VERIFIED source commit %s\n' "$sha"
  printf 'author: %s\ncommitter: %s\nsubject: %s\n' "$author" "$committer" "$subject"
}

verify_remote_commit() {
  local commit="$1" sha parents committer
  verify_https_origin
  sha="$(git rev-parse --verify "${commit}^{commit}" 2>/dev/null)" || fail 'fetched remote commit does not resolve'
  parents="$(parent_count "$sha")"
  if [[ "$parents" -gt 1 ]]; then
    committer="$(git show -s --format='%cn <%ce>' "$sha")"
    if [[ "$committer" == 'GitHub <noreply@github.com>' ]]; then
      printf 'GitHub-generated merge commit %s; not attested as a CloudBot source commit.\n' "$sha"
      return 0
    fi
    fail 'remote merge commit is not attested as a CloudBot source commit'
  fi
  verify_source_commit "$sha"
  printf 'VERIFIED remote commit %s\n' "$sha"
}

commit_source() {
  local message="$1" author_ident committer_ident
  verify_https_origin
  verify_subject "$message"
  if git diff --cached --quiet; then
    fail 'stage the intended source changes before using commit'
  fi

  # Restrict identity configuration to this checkout and neutralize inherited
  # identity overrides for the commit subprocess. Hooks remain enabled.
  git config --local user.name "$expected_name"
  git config --local user.email "$expected_email"
  author_ident="$(env -u GIT_AUTHOR_NAME -u GIT_AUTHOR_EMAIL git var GIT_AUTHOR_IDENT)"
  committer_ident="$(env -u GIT_COMMITTER_NAME -u GIT_COMMITTER_EMAIL git var GIT_COMMITTER_IDENT)"
  [[ "$author_ident" == "$expected_name <$expected_email>"* ]] || fail 'repository-local Git author identity did not take effect'
  [[ "$committer_ident" == "$expected_name <$expected_email>"* ]] || fail 'repository-local Git committer identity did not take effect'
  env -u GIT_AUTHOR_NAME -u GIT_AUTHOR_EMAIL -u GIT_COMMITTER_NAME -u GIT_COMMITTER_EMAIL \
    git commit -m "$message"
  verify_source_commit HEAD
}

publish_source() {
  local branch remote_line remote_sha head_sha
  verify_source_commit HEAD
  [[ -z "$(git status --porcelain --untracked-files=all)" ]] || fail 'working tree must be clean before publishing'
  branch="$(git symbolic-ref --quiet --short HEAD)" || fail 'publish requires a named branch'
  git check-ref-format "refs/heads/${branch}" >/dev/null || fail 'current branch name is not a valid push ref'
  remote_line="$(GIT_TERMINAL_PROMPT=0 git ls-remote --heads origin "refs/heads/${branch}")" || fail 'could not inspect the remote branch'

  if [[ -n "$remote_line" ]]; then
    remote_sha="${remote_line%%$'\t'*}"
    GIT_TERMINAL_PROMPT=0 git fetch --no-tags origin "refs/heads/${branch}"
    remote_sha="$(git rev-parse --verify FETCH_HEAD)"
    [[ "$remote_sha" == "${remote_line%%$'\t'*}" ]] || fail 'remote branch changed during fast-forward inspection; retry from a fresh fetch'
    git merge-base --is-ancestor "$remote_sha" HEAD || fail 'remote target is not an ancestor; refusing a non-fast-forward push'
  else
    GIT_TERMINAL_PROMPT=0 git fetch --no-tags origin refs/heads/main || fail 'could not fetch origin/main for new-branch ancestry validation'
    git merge-base --is-ancestor FETCH_HEAD HEAD || fail 'new branch must descend from origin/main'
  fi

  printf 'Dry-run push target: origin refs/heads/%s\n' "$branch"
  GIT_TERMINAL_PROMPT=0 git push --dry-run origin "HEAD:refs/heads/${branch}"
  GIT_TERMINAL_PROMPT=0 git push origin "HEAD:refs/heads/${branch}"
  GIT_TERMINAL_PROMPT=0 git fetch --no-tags origin "refs/heads/${branch}"
  head_sha="$(git rev-parse --verify HEAD)"
  remote_sha="$(git rev-parse --verify FETCH_HEAD)"
  [[ "$remote_sha" == "$head_sha" ]] || fail 'remote tip differs from the verified local source commit after push'
  verify_remote_commit FETCH_HEAD
  printf 'PUBLISHED verified source commit %s\n' "$head_sha"
}

main() {
  require_repo
  local command="${1:-}"
  case "$command" in
    commit)
      [[ "$#" == 2 ]] || { usage >&2; exit 2; }
      commit_source "$2"
      ;;
    check-source)
      [[ "$#" == 2 ]] || { usage >&2; exit 2; }
      verify_source_commit "$2"
      ;;
    check-remote)
      [[ "$#" == 2 ]] || { usage >&2; exit 2; }
      verify_remote_commit "$2"
      ;;
    publish)
      [[ "$#" == 1 ]] || { usage >&2; exit 2; }
      publish_source
      ;;
    -h|--help)
      usage
      ;;
    *)
      usage >&2
      exit 2
      ;;
  esac
}

main "$@"

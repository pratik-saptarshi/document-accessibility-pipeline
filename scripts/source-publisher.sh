#!/usr/bin/env bash
set -euo pipefail

readonly expected_name='CloudBot Neo-green'
readonly expected_email='waynemanorsaptarshi@gmail.com'
readonly expected_fetch_url='https://github.com/pratik-saptarshi/document-accessibility-pipeline.git'
readonly expected_fetch_url_no_suffix='https://github.com/pratik-saptarshi/document-accessibility-pipeline'
readonly conventional_subject='^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-z0-9][a-z0-9._/-]*\))?!?: [^[:space:]].*$'

usage() {
  cat <<'USAGE'
Usage:
  scripts/source-publisher.sh commit "type(scope): imperative summary"
  scripts/source-publisher.sh check-source <commit-ish>
  scripts/source-publisher.sh check-remote <fetched-ref>
  scripts/source-publisher.sh publish

commit sets repository-local identity, creates one Conventional Commit, and
checks its stored author and committer. publish verifies every outgoing commit
claiming the publisher identity, checks the fast-forward target, performs a
noninteractive HTTPS push dry run followed by a normal push, fetches the remote
tip, and checks the exact verified object.
USAGE
}

fail() {
  printf 'publisher check failed: %s\n' "$1" >&2
  exit 1
}

require_repo() {
  git_raw rev-parse --show-toplevel >/dev/null 2>&1 || fail 'run this command inside a Git checkout'
}

git_raw() {
  GIT_NO_REPLACE_OBJECTS=1 git "$@"
}

is_expected_repo_url() {
  [[ "$1" == "$expected_fetch_url" || "$1" == "$expected_fetch_url_no_suffix" ]]
}

verify_https_origin() {
  local fetch_list push_list url
  local -a fetch_urls push_urls
  fetch_list="$(git_raw remote get-url --all origin 2>/dev/null)" || fail 'origin fetch remote is missing'
  push_list="$(git_raw remote get-url --push --all origin 2>/dev/null)" || fail 'origin push remote is missing'
  mapfile -t fetch_urls <<<"$fetch_list"
  mapfile -t push_urls <<<"$push_list"
  [[ "${#fetch_urls[@]}" == 1 ]] || fail 'origin must have exactly one fetch destination'
  [[ "${#push_urls[@]}" == 1 ]] || fail 'origin must have exactly one push destination'
  for url in "${fetch_urls[0]}" "${push_urls[0]}"; do
    [[ "$url" == https://* ]] || fail 'origin fetch and push remotes must use HTTPS'
    [[ "$url" != *'@'* ]] || fail 'origin URLs must not embed credentials'
    is_expected_repo_url "$url" || fail 'origin must resolve only to the expected GitHub repository'
  done
}

verify_subject() {
  local subject="$1"
  [[ "$subject" =~ $conventional_subject ]] || fail 'source commit subject must follow Conventional Commits'
}

parent_count() {
  local line
  local -a fields
  line="$(git_raw rev-list --parents -n 1 "$1")" || fail 'could not resolve commit parents'
  read -r -a fields <<<"$line"
  printf '%s' "$(( ${#fields[@]} - 1 ))"
}

verify_source_commit() {
  local commit="$1" sha parents author committer subject
  verify_https_origin
  sha="$(git_raw rev-parse --verify "${commit}^{commit}" 2>/dev/null)" || fail 'source commit does not resolve'
  parents="$(parent_count "$sha")"
  [[ "$parents" == 1 ]] || fail 'agent-authored source commit must have exactly one parent; merge history is not re-authored'
  author="$(git_raw show -s --format='%an <%ae>' "$sha")"
  committer="$(git_raw show -s --format='%cn <%ce>' "$sha")"
  subject="$(git_raw show -s --format='%s' "$sha")"
  [[ "$author" == "${expected_name} <${expected_email}>" ]] || fail 'source commit author does not match the configured publisher identity'
  [[ "$committer" == "${expected_name} <${expected_email}>" ]] || fail 'source commit committer does not match the configured publisher identity'
  verify_subject "$subject"
  printf 'VERIFIED source commit %s\n' "$sha"
  printf 'author: %s\ncommitter: %s\nsubject: %s\n' "$author" "$committer" "$subject"
}

verify_remote_commit() {
  local commit="$1" sha parents committer
  verify_https_origin
  sha="$(git_raw rev-parse --verify "${commit}^{commit}" 2>/dev/null)" || fail 'fetched remote commit does not resolve'
  committer="$(git_raw show -s --format='%cn <%ce>' "$sha")"
  if [[ "$committer" == 'GitHub <noreply@github.com>' ]]; then
    printf 'UNVERIFIED GitHub-committer candidate %s; commit metadata alone is not authenticated as GitHub-generated and is not attested as a CloudBot source commit.\n' "$sha" >&2
    exit 3
  fi
  parents="$(parent_count "$sha")"
  if [[ "$parents" -gt 1 ]]; then
    fail 'remote merge commit is not attested as a CloudBot source commit'
  fi
  verify_source_commit "$sha"
  printf 'VERIFIED remote commit %s\n' "$sha"
}

verify_outgoing_commits() {
  local base="$1" target="$2" sha author_name author_email committer_name committer_email
  local -a commits
  mapfile -t commits < <(git_raw rev-list --reverse "${base}..${target}")
  for sha in "${commits[@]}"; do
    author_name="$(git_raw show -s --format='%an' "$sha")"
    author_email="$(git_raw show -s --format='%ae' "$sha")"
    committer_name="$(git_raw show -s --format='%cn' "$sha")"
    committer_email="$(git_raw show -s --format='%ce' "$sha")"
    if [[ "$committer_name" == GitHub && "$committer_email" == noreply@github.com ]]; then
      printf 'UNATTESTED GitHub-committer candidate in outgoing history: %s\n' "$sha"
      continue
    fi
    if [[ "$author_name" == "$expected_name" || "$author_email" == "$expected_email" || \
      "$committer_name" == "$expected_name" || "$committer_email" == "$expected_email" ]]; then
      verify_source_commit "$sha"
    else
      printf 'PRESERVED non-publisher commit in outgoing history: %s\n' "$sha"
    fi
  done
}

commit_source() {
  local message="$1" author_ident committer_ident
  verify_https_origin
  verify_subject "$message"
  if git_raw diff --cached --quiet; then
    fail 'stage the intended source changes before using commit'
  fi

  # Restrict identity configuration to this checkout and neutralize inherited
  # identity overrides for the commit subprocess. Hooks remain enabled.
  git config --local user.name "$expected_name"
  git config --local user.email "$expected_email"
  author_ident="$(env -u GIT_AUTHOR_NAME -u GIT_AUTHOR_EMAIL GIT_NO_REPLACE_OBJECTS=1 git var GIT_AUTHOR_IDENT)"
  committer_ident="$(env -u GIT_COMMITTER_NAME -u GIT_COMMITTER_EMAIL GIT_NO_REPLACE_OBJECTS=1 git var GIT_COMMITTER_IDENT)"
  [[ "$author_ident" == "$expected_name <$expected_email>"* ]] || fail 'repository-local Git author identity did not take effect'
  [[ "$committer_ident" == "$expected_name <$expected_email>"* ]] || fail 'repository-local Git committer identity did not take effect'
  env -u GIT_AUTHOR_NAME -u GIT_AUTHOR_EMAIL -u GIT_COMMITTER_NAME -u GIT_COMMITTER_EMAIL \
    GIT_NO_REPLACE_OBJECTS=1 git commit -m "$message"
  verify_source_commit HEAD
}

publish_source() {
  local branch remote_line remote_sha verified_sha current_head
  verified_sha="$(git_raw rev-parse --verify HEAD)" || fail 'could not capture the source commit to publish'
  verify_source_commit "$verified_sha"
  [[ -z "$(git_raw status --porcelain --untracked-files=all)" ]] || fail 'working tree must be clean before publishing'
  branch="$(git_raw symbolic-ref --quiet --short HEAD)" || fail 'publish requires a named branch'
  git_raw check-ref-format "refs/heads/${branch}" >/dev/null || fail 'current branch name is not a valid push ref'
  current_head="$(git_raw rev-parse --verify HEAD)"
  [[ "$current_head" == "$verified_sha" ]] || fail 'HEAD moved during source verification; aborting before network access'
  remote_line="$(GIT_TERMINAL_PROMPT=0 GIT_NO_REPLACE_OBJECTS=1 git ls-remote --heads origin "refs/heads/${branch}")" || fail 'could not inspect the remote branch'

  if [[ -n "$remote_line" ]]; then
    remote_sha="${remote_line%%$'\t'*}"
    GIT_TERMINAL_PROMPT=0 GIT_NO_REPLACE_OBJECTS=1 git fetch --no-tags origin "refs/heads/${branch}"
    remote_sha="$(git_raw rev-parse --verify FETCH_HEAD)"
    [[ "$remote_sha" == "${remote_line%%$'\t'*}" ]] || fail 'remote branch changed during fast-forward inspection; retry from a fresh fetch'
    git_raw merge-base --is-ancestor "$remote_sha" "$verified_sha" || fail 'remote target is not an ancestor; refusing a non-fast-forward push'
    verify_outgoing_commits "$remote_sha" "$verified_sha"
  else
    GIT_TERMINAL_PROMPT=0 GIT_NO_REPLACE_OBJECTS=1 git fetch --no-tags origin refs/heads/main || fail 'could not fetch origin/main for new-branch ancestry validation'
    remote_sha="$(git_raw rev-parse --verify FETCH_HEAD)"
    git_raw merge-base --is-ancestor "$remote_sha" "$verified_sha" || fail 'new branch must descend from origin/main'
    verify_outgoing_commits "$remote_sha" "$verified_sha"
  fi

  printf 'Dry-run push target: origin refs/heads/%s\n' "$branch"
  GIT_TERMINAL_PROMPT=0 GIT_NO_REPLACE_OBJECTS=1 git push --dry-run origin "${verified_sha}:refs/heads/${branch}"
  [[ "$(git_raw rev-parse --verify HEAD)" == "$verified_sha" ]] || fail 'HEAD moved during push dry run; aborting before the ordinary push'
  GIT_TERMINAL_PROMPT=0 GIT_NO_REPLACE_OBJECTS=1 git push origin "${verified_sha}:refs/heads/${branch}"
  GIT_TERMINAL_PROMPT=0 GIT_NO_REPLACE_OBJECTS=1 git fetch --no-tags origin "refs/heads/${branch}"
  remote_sha="$(git_raw rev-parse --verify FETCH_HEAD)"
  [[ "$remote_sha" == "$verified_sha" ]] || fail 'remote tip differs from the verified source object after push'
  verify_remote_commit FETCH_HEAD
  printf 'PUBLISHED verified source commit %s\n' "$verified_sha"
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

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
  scripts/source-publisher.sh merge-main "chore(scope): merge current main"
  scripts/source-publisher.sh check-source <commit-ish>
  scripts/source-publisher.sh check-remote <fetched-ref>
  scripts/source-publisher.sh publish

commit sets repository-local identity, creates one Conventional Commit, and
checks its stored author and committer. merge-main creates a local two-parent
merge of fetched origin/main with process-scoped publisher identity. publish
verifies every outgoing commit claiming the publisher identity, checks the
fast-forward target, performs a noninteractive HTTPS push dry run followed by
a normal push, fetches the remote tip, and checks the exact verified object.
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
  local fetch_list push_list url mirror_status
  local -a fetch_urls push_urls
  fetch_list="$(git_raw remote get-url --all origin 2>/dev/null)" || fail 'origin fetch remote is missing'
  push_list="$(git_raw remote get-url --push --all origin 2>/dev/null)" || fail 'origin push remote is missing'
  mapfile -t fetch_urls <<<"$fetch_list"
  mapfile -t push_urls <<<"$push_list"
  [[ "${#fetch_urls[@]}" == 1 ]] || fail 'origin must have exactly one fetch destination'
  [[ "${#push_urls[@]}" == 1 ]] || fail 'origin must have exactly one push destination'
  if mirror_status="$(git_raw config --bool --get remote.origin.mirror 2>/dev/null)"; then
    [[ "$mirror_status" != true ]] || fail 'origin mirror mode is not supported for source publication'
  else
    [[ "$?" == 1 ]] || fail 'could not validate origin mirror mode'
  fi
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
  [[ "$parents" == 1 || "$parents" == 2 ]] || fail 'source commits must have one parent, or two parents for a main-sync merge'
  author="$(git_raw show -s --format='%an <%ae>' "$sha")"
  committer="$(git_raw show -s --format='%cn <%ce>' "$sha")"
  subject="$(git_raw show -s --format='%s' "$sha")"
  [[ "$author" == "${expected_name} <${expected_email}>" ]] || fail 'source commit author does not match the configured publisher identity'
  [[ "$committer" == "${expected_name} <${expected_email}>" ]] || fail 'source commit committer does not match the configured publisher identity'
  verify_subject "$subject"
  printf 'VERIFIED publisher identity and subject for %s\n' "$sha"
  printf 'author: %s\ncommitter: %s\nsubject: %s\n' "$author" "$committer" "$subject"
}

fetch_main_sha() {
  GIT_TERMINAL_PROMPT=0 GIT_NO_REPLACE_OBJECTS=1 git fetch --no-tags origin refs/heads/main || \
    fail 'could not fetch origin/main for source verification'
  git_raw rev-parse --verify FETCH_HEAD
}

verify_main_sync_merge() {
  local commit="$1" main_sha="$2" feature_base="${3:-}" line first_parent second_parent
  local -a fields
  line="$(git_raw rev-list --parents -n 1 "$commit")" || fail 'could not resolve main-sync merge parents'
  read -r -a fields <<<"$line"
  [[ "${#fields[@]}" == 3 ]] || fail 'publisher main-sync merge must have exactly two parents'
  first_parent="${fields[1]}"
  second_parent="${fields[2]}"
  git_raw merge-base --is-ancestor "$second_parent" "$main_sha" || \
    fail 'publisher merge second parent is not part of fetched origin/main'
  if [[ -n "$feature_base" ]]; then
    git_raw merge-base --is-ancestor "$feature_base" "$first_parent" || \
      fail 'publisher merge first parent does not preserve the published feature branch'
  fi
}

verify_source_candidate() {
  local commit="$1" main_sha="${2:-}" feature_base="${3:-}" parents
  verify_source_commit "$commit"
  parents="$(parent_count "$commit")"
  if [[ "$parents" == 2 ]]; then
    [[ -n "$main_sha" ]] || main_sha="$(fetch_main_sha)"
    verify_main_sync_merge "$commit" "$main_sha" "$feature_base"
  fi
  printf 'VERIFIED source commit %s\n' "$(git_raw rev-parse --verify "${commit}^{commit}")"
}

verify_remote_commit() {
  local commit="$1" sha committer
  verify_https_origin
  sha="$(git_raw rev-parse --verify "${commit}^{commit}" 2>/dev/null)" || fail 'fetched remote commit does not resolve'
  committer="$(git_raw show -s --format='%cn <%ce>' "$sha")"
  if [[ "$committer" == 'GitHub <noreply@github.com>' ]]; then
    printf 'UNVERIFIED GitHub-committer candidate %s; commit metadata alone is not authenticated as GitHub-generated and is not attested as a CloudBot source commit.\n' "$sha" >&2
    exit 3
  fi
  verify_source_candidate "$sha" "${2:-}"
  printf 'VERIFIED remote commit %s\n' "$sha"
}

verify_outgoing_commits() {
  local base="$1" target="$2" main_sha="$3" feature_base="${4:-}" sha author_name author_email committer_name committer_email commit_list
  local -a commits=()
  # Main is a fetched trusted remote, so commits already reachable from it
  # are not new feature-branch source commits in this publication.
  commit_list="$(git_raw rev-list --reverse "${base}..${target}" --not "$main_sha")" || fail 'could not enumerate outgoing commits'
  if [[ -n "$commit_list" ]]; then
    mapfile -t commits <<<"$commit_list"
  fi
  for sha in "${commits[@]}"; do
    author_name="$(git_raw show -s --format='%an' "$sha")"
    author_email="$(git_raw show -s --format='%ae' "$sha")"
    committer_name="$(git_raw show -s --format='%cn' "$sha")"
    committer_email="$(git_raw show -s --format='%ce' "$sha")"
    if [[ "$author_name" == "$expected_name" || "$author_email" == "$expected_email" || \
      "$committer_name" == "$expected_name" || "$committer_email" == "$expected_email" ]]; then
      verify_source_candidate "$sha" "$main_sha" "$feature_base"
    elif [[ "$committer_name" == GitHub && "$committer_email" == noreply@github.com ]]; then
      printf 'UNATTESTED GitHub-committer candidate in outgoing history: %s\n' "$sha"
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

merge_main() {
  local message="$1" branch current_head main_sha
  verify_https_origin
  verify_subject "$message"
  [[ -z "$(git_raw status --porcelain --untracked-files=all)" ]] || fail 'worktree must be clean before merging origin/main'
  branch="$(git_raw symbolic-ref --quiet --short HEAD)" || fail 'merge-main requires a named branch'
  [[ "$branch" != main ]] || fail 'merge-main cannot merge origin/main into main'
  git_raw check-ref-format "refs/heads/${branch}" >/dev/null || fail 'current branch name is not a valid ref'
  current_head="$(git_raw rev-parse --verify HEAD)"
  main_sha="$(fetch_main_sha)"
  [[ "$(git_raw rev-parse --verify HEAD)" == "$current_head" ]] || fail 'HEAD moved while fetching origin/main; retry from the captured branch tip'
  if git_raw merge-base --is-ancestor "$main_sha" "$current_head"; then
    printf 'origin/main %s is already included in %s; no merge commit created\n' "$main_sha" "$branch"
    return 0
  fi
  GIT_AUTHOR_NAME="$expected_name" GIT_AUTHOR_EMAIL="$expected_email" \
  GIT_COMMITTER_NAME="$expected_name" GIT_COMMITTER_EMAIL="$expected_email" \
  GIT_NO_REPLACE_OBJECTS=1 git merge --no-ff --no-edit -m "$message" "$main_sha" || \
    fail 'local origin/main merge failed; resolve conflicts before publishing'
  verify_source_candidate HEAD "$main_sha" "$current_head"
}

publish_source() {
  local branch remote_line remote_sha verified_sha current_head main_sha feature_base=''
  verified_sha="$(git_raw rev-parse --verify HEAD)" || fail 'could not capture the source commit to publish'
  verify_source_commit "$verified_sha"
  [[ -z "$(git_raw status --porcelain --untracked-files=all)" ]] || fail 'working tree must be clean before publishing'
  branch="$(git_raw symbolic-ref --quiet --short HEAD)" || fail 'publish requires a named branch'
  git_raw check-ref-format "refs/heads/${branch}" >/dev/null || fail 'current branch name is not a valid push ref'
  main_sha="$(fetch_main_sha)"
  current_head="$(git_raw rev-parse --verify HEAD)"
  [[ "$current_head" == "$verified_sha" ]] || fail 'HEAD moved during source verification; aborting before network access'
  remote_line="$(GIT_TERMINAL_PROMPT=0 GIT_NO_REPLACE_OBJECTS=1 git ls-remote --heads origin "refs/heads/${branch}")" || fail 'could not inspect the remote branch'

  if [[ -n "$remote_line" ]]; then
    remote_sha="${remote_line%%$'\t'*}"
    GIT_TERMINAL_PROMPT=0 GIT_NO_REPLACE_OBJECTS=1 git fetch --no-tags origin "refs/heads/${branch}"
    remote_sha="$(git_raw rev-parse --verify FETCH_HEAD)"
    [[ "$remote_sha" == "${remote_line%%$'\t'*}" ]] || fail 'remote branch changed during fast-forward inspection; retry from a fresh fetch'
    git_raw merge-base --is-ancestor "$remote_sha" "$verified_sha" || fail 'remote target is not an ancestor; refusing a non-fast-forward push'
    feature_base="$remote_sha"
  else
    remote_sha="$main_sha"
    git_raw merge-base --is-ancestor "$remote_sha" "$verified_sha" || fail 'new branch must descend from origin/main'
  fi
  verify_outgoing_commits "$remote_sha" "$verified_sha" "$main_sha" "$feature_base"

  printf 'Dry-run push target: origin refs/heads/%s\n' "$branch"
  GIT_TERMINAL_PROMPT=0 GIT_NO_REPLACE_OBJECTS=1 git push --dry-run --no-follow-tags --no-mirror origin "${verified_sha}:refs/heads/${branch}"
  [[ "$(git_raw rev-parse --verify HEAD)" == "$verified_sha" ]] || fail 'HEAD moved during push dry run; aborting before the ordinary push'
  GIT_TERMINAL_PROMPT=0 GIT_NO_REPLACE_OBJECTS=1 git push --no-follow-tags --no-mirror origin "${verified_sha}:refs/heads/${branch}"
  GIT_TERMINAL_PROMPT=0 GIT_NO_REPLACE_OBJECTS=1 git fetch --no-tags origin "refs/heads/${branch}"
  remote_sha="$(git_raw rev-parse --verify FETCH_HEAD)"
  [[ "$remote_sha" == "$verified_sha" ]] || fail 'remote tip differs from the verified source object after push'
  verify_remote_commit FETCH_HEAD "$main_sha"
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
    merge-main)
      [[ "$#" == 2 ]] || { usage >&2; exit 2; }
      merge_main "$2"
      ;;
    check-source)
      [[ "$#" == 2 ]] || { usage >&2; exit 2; }
      verify_source_candidate "$2"
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

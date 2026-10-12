#!/usr/bin/env bash
set -euo pipefail

repo_root="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)"
publisher="${repo_root}/scripts/source-publisher.sh"
test_root="$(mktemp -d)"
trap 'rm -rf -- "${test_root}"' EXIT

export GIT_CONFIG_NOSYSTEM=1
export GIT_CONFIG_GLOBAL=/dev/null
export GIT_TERMINAL_PROMPT=0

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  exit 1
}

expect_rejected() {
  local label="$1"
  shift
  if "$@" >"${test_root}/stdout" 2>"${test_root}/stderr"; then
    fail "${label} was unexpectedly accepted"
  fi
}

expect_hosted_candidate() {
  local label="$1" output status
  shift
  set +e
  output="$("$@" 2>&1)"
  status=$?
  set -e
  [[ "$status" == 3 ]] || fail "${label} returned ${status}, want explicit unverified status 3"
  [[ "$output" == *'GitHub-committer candidate'* && "$output" == *'not authenticated as GitHub-generated'* ]] || \
    fail "${label} was not clearly identified as an unverified hosted candidate"
}

new_repo() {
  local name="$1"
  local path="${test_root}/${name}"
  mkdir -p "${path}"
  git -C "${path}" init --quiet --initial-branch=main
  git -C "${path}" config user.name 'Human Contributor'
  git -C "${path}" config user.email 'human@example.invalid'
  git -C "${path}" remote add origin 'https://github.com/pratik-saptarshi/document-accessibility-pipeline.git'
  git -C "${path}" commit --quiet --allow-empty -m 'chore(test): create isolated test repository'
  printf '%s' "${path}"
}

valid="$(new_repo valid)"
git -C "${valid}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): valid publisher identity'
(
  cd "${valid}"
  "${publisher}" check-source HEAD
)

git -C "${valid}" update-ref refs/remotes/origin/publisher-test HEAD
(
  cd "${valid}"
  "${publisher}" check-remote refs/remotes/origin/publisher-test
)

wrapped_commit="$(new_repo wrapped-commit)"
git -C "${wrapped_commit}" config --local user.name 'Human Contributor'
git -C "${wrapped_commit}" config --local user.email 'human@example.invalid'
printf 'synthetic fixture\n' >"${wrapped_commit}/fixture.txt"
git -C "${wrapped_commit}" add fixture.txt
(
  cd "${wrapped_commit}"
  GIT_AUTHOR_NAME='Inherited Author' GIT_AUTHOR_EMAIL='inherited-author@example.invalid' \
  GIT_COMMITTER_NAME='Inherited Committer' GIT_COMMITTER_EMAIL='inherited-committer@example.invalid' \
    "${publisher}" commit 'docs(test): exercise safe commit wrapper'
)
[[ "$(git -C "${wrapped_commit}" config --local user.name)" == 'CloudBot Neo-green' ]] || fail 'commit wrapper did not set repo-local author name'
[[ "$(git -C "${wrapped_commit}" config --local user.email)" == 'waynemanorsaptarshi@gmail.com' ]] || fail 'commit wrapper did not set repo-local author email'
[[ "$(git -C "${wrapped_commit}" show -s --format='%an <%ae>' HEAD)" == 'CloudBot Neo-green <waynemanorsaptarshi@gmail.com>' ]] || fail 'commit wrapper author metadata is incorrect'
[[ "$(git -C "${wrapped_commit}" show -s --format='%cn <%ce>' HEAD)" == 'CloudBot Neo-green <waynemanorsaptarshi@gmail.com>' ]] || fail 'commit wrapper committer metadata is incorrect'

bad_author="$(new_repo bad-author)"
GIT_AUTHOR_NAME='Different Author' GIT_AUTHOR_EMAIL='different@example.invalid' \
  git -C "${bad_author}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): wrong author'
expect_rejected 'wrong author' bash -c "cd '${bad_author}' && '${publisher}' check-source HEAD"

bad_committer="$(new_repo bad-committer)"
GIT_COMMITTER_NAME='Different Committer' GIT_COMMITTER_EMAIL='different@example.invalid' \
  git -C "${bad_committer}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): wrong committer'
expect_rejected 'wrong committer' bash -c "cd '${bad_committer}' && '${publisher}' check-source HEAD"

bad_subject="$(new_repo bad-subject)"
git -C "${bad_subject}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'This is not conventional'
expect_rejected 'non-Conventional subject' bash -c "cd '${bad_subject}' && '${publisher}' check-source HEAD"

bad_remote="$(new_repo bad-remote)"
git -C "${bad_remote}" remote set-url --push origin 'git@github.com:pratik-saptarshi/document-accessibility-pipeline.git'
git -C "${bad_remote}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): reject ssh publishing'
expect_rejected 'SSH push URL' bash -c "cd '${bad_remote}' && '${publisher}' check-source HEAD"

credential_remote="$(new_repo credential-remote)"
git -C "${credential_remote}" remote set-url --push origin \
  'https://user:synthetic-secret@example.invalid/repo.git'
git -C "${credential_remote}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): reject embedded credentials'
expect_rejected 'credential-bearing push URL' bash -c "cd '${credential_remote}' && '${publisher}' check-source HEAD"
if rg -q 'synthetic-secret' "${test_root}/stdout" "${test_root}/stderr"; then
  fail 'credential-bearing URL leaked in diagnostics'
fi

wrong_repository="$(new_repo wrong-repository)"
git -C "${wrong_repository}" remote set-url --push origin \
  'https://github.com/example/other-repository.git'
git -C "${wrong_repository}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): reject a different repository'
expect_rejected 'wrong repository push URL' bash -c "cd '${wrong_repository}' && '${publisher}' check-source HEAD"

multiple_pushes="$(new_repo multiple-pushes)"
git -C "${multiple_pushes}" remote set-url --add --push origin \
  'https://github.com/pratik-saptarshi/document-accessibility-pipeline.git'
git -C "${multiple_pushes}" remote set-url --add --push origin \
  'git@github.com:attacker/other-repository.git'
git -C "${multiple_pushes}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): reject multiple push destinations'
expect_rejected 'multiple push destinations' bash -c "cd '${multiple_pushes}' && '${publisher}' check-source HEAD"

multiple_fetches="$(new_repo multiple-fetches)"
git -C "${multiple_fetches}" remote set-url --add origin \
  'https://github.com/example/other-repository.git'
git -C "${multiple_fetches}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): reject multiple fetch destinations'
expect_rejected 'multiple fetch destinations' bash -c "cd '${multiple_fetches}' && '${publisher}' check-source HEAD"

multiple_push_credentials="$(new_repo multiple-push-credentials)"
git -C "${multiple_push_credentials}" remote set-url --add --push origin \
  'https://github.com/pratik-saptarshi/document-accessibility-pipeline.git'
git -C "${multiple_push_credentials}" remote set-url --add --push origin \
  'https://user:second-synthetic-secret@example.invalid/repo.git'
git -C "${multiple_push_credentials}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): reject secondary credential destination'
expect_rejected 'secondary credential push destination' bash -c "cd '${multiple_push_credentials}' && '${publisher}' check-source HEAD"
if rg -q 'second-synthetic-secret' "${test_root}/stdout" "${test_root}/stderr"; then
  fail 'secondary credential-bearing URL leaked in diagnostics'
fi

merge="$(new_repo github-merge)"
git -C "${merge}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): second parent'
parent_one="$(git -C "${merge}" rev-parse HEAD)"
git -C "${merge}" checkout --quiet --detach HEAD^
tree="$(git -C "${merge}" rev-parse HEAD^{tree})"
parent_two="${parent_one}"
merge_sha="$(
  GIT_AUTHOR_NAME='Pull Request Author' GIT_AUTHOR_EMAIL='author@example.invalid' \
  GIT_COMMITTER_NAME='GitHub' GIT_COMMITTER_EMAIL='noreply@github.com' \
  git -C "${merge}" commit-tree "${tree}" -p "$(git -C "${merge}" rev-parse HEAD)" -p "${parent_two}" \
    -m 'Merge pull request #1 from example/branch'
)"
git -C "${merge}" update-ref refs/remotes/origin/main "${merge_sha}"
expect_hosted_candidate 'GitHub merge commit' bash -c "cd '${merge}' && '${publisher}' check-remote refs/remotes/origin/main"

squash="$(new_repo github-squash)"
base="$(git -C "${squash}" rev-parse HEAD)"
tree="$(git -C "${squash}" rev-parse HEAD^{tree})"
squash_sha="$(
  GIT_AUTHOR_NAME='Pull Request Author' GIT_AUTHOR_EMAIL='author@example.invalid' \
  GIT_COMMITTER_NAME='GitHub' GIT_COMMITTER_EMAIL='noreply@github.com' \
  git -C "${squash}" commit-tree "${tree}" -p "${base}" -m 'docs(test): hosted squash merge (#1)'
)"
git -C "${squash}" update-ref refs/remotes/origin/main "${squash_sha}"
expect_hosted_candidate 'GitHub squash commit' bash -c "cd '${squash}' && '${publisher}' check-remote refs/remotes/origin/main"

raw_identity="$(new_repo raw-identity)"
parent="$(git -C "${raw_identity}" rev-parse HEAD)"
tree="$(git -C "${raw_identity}" rev-parse HEAD^{tree})"
raw_sha="$(
  GIT_AUTHOR_NAME='Wrong Raw Author' GIT_AUTHOR_EMAIL='wrong@example.invalid' \
  GIT_COMMITTER_NAME='Wrong Raw Committer' GIT_COMMITTER_EMAIL='wrong@example.invalid' \
  git -C "${raw_identity}" commit-tree "${tree}" -p "${parent}" -m 'fix(test): forged replacement view'
)"
replacement_sha="$(
  GIT_AUTHOR_NAME='CloudBot Neo-green' GIT_AUTHOR_EMAIL='waynemanorsaptarshi@gmail.com' \
  GIT_COMMITTER_NAME='CloudBot Neo-green' GIT_COMMITTER_EMAIL='waynemanorsaptarshi@gmail.com' \
  git -C "${raw_identity}" commit-tree "${tree}" -p "${parent}" -m 'fix(test): replacement metadata must not mask raw object'
)"
git -C "${raw_identity}" update-ref refs/heads/main "${raw_sha}"
git -C "${raw_identity}" replace "${raw_sha}" "${replacement_sha}"
expect_rejected 'replace-ref-masked raw identity' bash -c "cd '${raw_identity}' && '${publisher}' check-source HEAD"

fake_bin="${test_root}/fake-bin"
mkdir -p "${fake_bin}"
real_git="$(command -v git)"
cat >"${fake_bin}/git" <<'GIT_SHIM'
#!/usr/bin/env bash
set -euo pipefail
case "${1:-}:${2:-}" in
  ls-remote:--heads)
    ref="${4}"
    if [[ "${ref}" == refs/heads/main ]]; then
      printf '%s\t%s\n' "${PUBLISHER_REMOTE_MAIN_SHA}" "${ref}"
    elif [[ -s "${PUBLISHER_REMOTE_BRANCH_FILE}" ]]; then
      printf '%s\t%s\n' "$(cat "${PUBLISHER_REMOTE_BRANCH_FILE}")" "${ref}"
    fi
    ;;
  fetch:--no-tags)
    ref="${4#refs/heads/}"
    if [[ "${ref}" == main ]]; then
      sha="${PUBLISHER_REMOTE_MAIN_SHA}"
    else
      sha="$(cat "${PUBLISHER_REMOTE_BRANCH_FILE}")"
    fi
    count=0
    [[ ! -f "${PUBLISHER_FETCH_COUNT_FILE}" ]] || count="$(cat "${PUBLISHER_FETCH_COUNT_FILE}")"
    count=$((count + 1))
    printf '%s\n' "${count}" >"${PUBLISHER_FETCH_COUNT_FILE}"
    [[ "${PUBLISHER_FAIL_FETCH:-0}" != 1 ]] || exit 41
    if [[ "${PUBLISHER_FAIL_POST_FETCH:-0}" == 1 && "${ref}" != main ]]; then exit 42; fi
    if [[ "${PUBLISHER_MISMATCH_READBACK:-0}" == 1 && "${ref}" != main ]]; then
      sha="${PUBLISHER_REMOTE_MAIN_SHA}"
    fi
    git_dir="$("${PUBLISHER_REAL_GIT}" rev-parse --git-dir)"
    printf '%s\t\tbranch %s of origin\n' "${sha}" "${ref}" >"${git_dir}/FETCH_HEAD"
    ;;
  push:--dry-run)
    printf 'dry-run:%s\n' "${4}" >>"${PUBLISHER_PUSH_LOG}"
    [[ "${PUBLISHER_FAIL_DRY_RUN:-0}" != 1 ]] || exit 43
    if [[ "${PUBLISHER_MOVE_HEAD_AFTER_DRY_RUN:-0}" == 1 ]]; then
      head_ref="$("${PUBLISHER_REAL_GIT}" symbolic-ref HEAD)"
      "${PUBLISHER_REAL_GIT}" update-ref "${head_ref}" "${PUBLISHER_MOVED_HEAD_SHA}"
    fi
    ;;
  push:origin)
    printf 'push:%s\n' "${3}" >>"${PUBLISHER_PUSH_LOG}"
    [[ "${PUBLISHER_FAIL_PUSH:-0}" != 1 ]] || exit 44
    refspec="${3}"
    sha="$("${PUBLISHER_REAL_GIT}" rev-parse --verify "${refspec%%:*}")"
    printf '%s\n' "${sha}" >"${PUBLISHER_REMOTE_BRANCH_FILE}"
    ;;
  *)
    exec "${PUBLISHER_REAL_GIT}" "$@"
    ;;
esac
GIT_SHIM
chmod 0755 "${fake_bin}/git"

push_repo="$(new_repo publish)"
base_sha="$(git -C "${push_repo}" rev-parse HEAD)"
git -C "${push_repo}" switch --quiet --create publisher-test
git -C "${push_repo}" -c user.name='Human Contributor' -c user.email='human@example.invalid' \
  commit --quiet --allow-empty -m 'Unchanged human contribution with legacy subject'
git -C "${push_repo}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): publish verified source commit'
export PUBLISHER_REAL_GIT="${real_git}"
export PUBLISHER_REMOTE_MAIN_SHA="${base_sha}"
export PUBLISHER_REMOTE_BRANCH_FILE="${test_root}/remote-branch-sha"
export PUBLISHER_PUSH_LOG="${test_root}/push.log"
export PUBLISHER_FETCH_COUNT_FILE="${test_root}/fetch-count"
export PATH="${fake_bin}:${PATH}"
(
  cd "${push_repo}"
  "${publisher}" publish
)
expected_source="$(git -C "${push_repo}" rev-parse HEAD)"
expected_refspec="${expected_source}:refs/heads/publisher-test"
[[ "$(cat "${PUBLISHER_PUSH_LOG}")" == "dry-run:${expected_refspec}"$'\n'"push:${expected_refspec}" ]] || \
  fail 'publisher did not dry-run then push the exact verified object'
[[ "$(cat "${PUBLISHER_REMOTE_BRANCH_FILE}")" == "${expected_source}" ]] || fail 'remote verification did not observe the pushed tip'

bad_outgoing="$(new_repo bad-outgoing-agent)"
base_sha="$(git -C "${bad_outgoing}" rev-parse HEAD)"
git -C "${bad_outgoing}" switch --quiet --create publisher-test
GIT_COMMITTER_NAME='Different Committer' GIT_COMMITTER_EMAIL='different@example.invalid' \
  git -C "${bad_outgoing}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
    commit --quiet --allow-empty -m 'fix(test): reject earlier bad agent commit'
git -C "${bad_outgoing}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): valid tip must not hide earlier bad agent commit'
export PUBLISHER_REMOTE_MAIN_SHA="${base_sha}"
export PUBLISHER_REMOTE_BRANCH_FILE="${test_root}/bad-outgoing-remote-sha"
export PUBLISHER_PUSH_LOG="${test_root}/bad-outgoing-push.log"
expect_rejected 'invalid earlier outgoing agent commit' bash -c "cd '${bad_outgoing}' && '${publisher}' publish"
[[ ! -s "${PUBLISHER_PUSH_LOG}" ]] || fail 'publisher attempted a push containing an invalid agent commit'

failed_dry_run="$(new_repo failed-dry-run)"
base_sha="$(git -C "${failed_dry_run}" rev-parse HEAD)"
git -C "${failed_dry_run}" switch --quiet --create publisher-test
git -C "${failed_dry_run}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): stop when dry run fails'
export PUBLISHER_REMOTE_MAIN_SHA="${base_sha}"
export PUBLISHER_REMOTE_BRANCH_FILE="${test_root}/failed-dry-run-remote-sha"
export PUBLISHER_PUSH_LOG="${test_root}/failed-dry-run-push.log"
export PUBLISHER_FETCH_COUNT_FILE="${test_root}/failed-dry-run-fetch-count"
export PUBLISHER_FAIL_DRY_RUN=1
expect_rejected 'failed push dry run' bash -c "cd '${failed_dry_run}' && '${publisher}' publish"
[[ "$(cat "${PUBLISHER_PUSH_LOG}")" == 'dry-run:'* ]] || fail 'publisher attempted ordinary push after failed dry run'
unset PUBLISHER_FAIL_DRY_RUN

failed_push="$(new_repo failed-push)"
base_sha="$(git -C "${failed_push}" rev-parse HEAD)"
git -C "${failed_push}" switch --quiet --create publisher-test
git -C "${failed_push}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): report failed ordinary push'
export PUBLISHER_REMOTE_MAIN_SHA="${base_sha}"
export PUBLISHER_REMOTE_BRANCH_FILE="${test_root}/failed-push-remote-sha"
export PUBLISHER_PUSH_LOG="${test_root}/failed-push.log"
export PUBLISHER_FETCH_COUNT_FILE="${test_root}/failed-push-fetch-count"
export PUBLISHER_FAIL_PUSH=1
expect_rejected 'failed ordinary push' bash -c "cd '${failed_push}' && '${publisher}' publish"
[[ "$(cat "${PUBLISHER_PUSH_LOG}")" == $'dry-run:'* && "$(cat "${PUBLISHER_PUSH_LOG}")" == *$'\npush:'* ]] || fail 'ordinary push was not attempted after successful dry run'
unset PUBLISHER_FAIL_PUSH

post_fetch_failure="$(new_repo post-fetch-failure)"
base_sha="$(git -C "${post_fetch_failure}" rev-parse HEAD)"
git -C "${post_fetch_failure}" switch --quiet --create publisher-test
git -C "${post_fetch_failure}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): report failed post-push fetch'
export PUBLISHER_REMOTE_MAIN_SHA="${base_sha}"
export PUBLISHER_REMOTE_BRANCH_FILE="${test_root}/post-fetch-failure-remote-sha"
export PUBLISHER_PUSH_LOG="${test_root}/post-fetch-failure.log"
export PUBLISHER_FETCH_COUNT_FILE="${test_root}/post-fetch-failure-fetch-count"
export PUBLISHER_FAIL_POST_FETCH=1
expect_rejected 'failed post-push fetch' bash -c "cd '${post_fetch_failure}' && '${publisher}' publish"
[[ "$(cat "${PUBLISHER_PUSH_LOG}")" == $'dry-run:'* && "$(cat "${PUBLISHER_PUSH_LOG}")" == *$'\npush:'* ]] || fail 'post-push fetch failure did not occur after the push'
unset PUBLISHER_FAIL_POST_FETCH

failed_fetch="$(new_repo failed-fetch)"
base_sha="$(git -C "${failed_fetch}" rev-parse HEAD)"
git -C "${failed_fetch}" switch --quiet --create publisher-test
git -C "${failed_fetch}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): report failed ancestry fetch'
export PUBLISHER_REMOTE_MAIN_SHA="${base_sha}"
export PUBLISHER_REMOTE_BRANCH_FILE="${test_root}/failed-fetch-remote-sha"
export PUBLISHER_PUSH_LOG="${test_root}/failed-fetch.log"
export PUBLISHER_FETCH_COUNT_FILE="${test_root}/failed-fetch-count"
export PUBLISHER_FAIL_FETCH=1
expect_rejected 'failed base fetch' bash -c "cd '${failed_fetch}' && '${publisher}' publish"
[[ ! -s "${PUBLISHER_PUSH_LOG}" ]] || fail 'publisher attempted push after ancestry fetch failed'
unset PUBLISHER_FAIL_FETCH

mismatched_readback="$(new_repo mismatched-readback)"
base_sha="$(git -C "${mismatched_readback}" rev-parse HEAD)"
git -C "${mismatched_readback}" switch --quiet --create publisher-test
git -C "${mismatched_readback}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): reject mismatched remote readback'
export PUBLISHER_REMOTE_MAIN_SHA="${base_sha}"
export PUBLISHER_REMOTE_BRANCH_FILE="${test_root}/mismatched-readback-remote-sha"
export PUBLISHER_PUSH_LOG="${test_root}/mismatched-readback.log"
export PUBLISHER_FETCH_COUNT_FILE="${test_root}/mismatched-readback-fetch-count"
export PUBLISHER_MISMATCH_READBACK=1
expect_rejected 'mismatched remote readback' bash -c "cd '${mismatched_readback}' && '${publisher}' publish"
unset PUBLISHER_MISMATCH_READBACK

moved_head="$(new_repo moved-head)"
base_sha="$(git -C "${moved_head}" rev-parse HEAD)"
git -C "${moved_head}" switch --quiet --create publisher-test
git -C "${moved_head}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): verified commit before head movement'
tree="$(git -C "${moved_head}" rev-parse HEAD^{tree})"
PUBLISHER_MOVED_HEAD_SHA="$(git -C "${moved_head}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit-tree "${tree}" -p "${base_sha}" -m 'fix(test): unexpected commit after dry run')"
export PUBLISHER_REMOTE_MAIN_SHA="${base_sha}"
export PUBLISHER_REMOTE_BRANCH_FILE="${test_root}/moved-head-remote-sha"
export PUBLISHER_PUSH_LOG="${test_root}/moved-head.log"
export PUBLISHER_FETCH_COUNT_FILE="${test_root}/moved-head-fetch-count"
export PUBLISHER_MOVED_HEAD_SHA
export PUBLISHER_MOVE_HEAD_AFTER_DRY_RUN=1
expect_rejected 'HEAD movement after dry run' bash -c "cd '${moved_head}' && '${publisher}' publish"
[[ "$(cat "${PUBLISHER_PUSH_LOG}")" == 'dry-run:'* ]] || fail 'publisher pushed after HEAD moved during the dry run'
[[ "$(git -C "${moved_head}" rev-parse HEAD)" == "${PUBLISHER_MOVED_HEAD_SHA}" ]] || fail 'HEAD movement test fixture was not applied'
[[ -z "$(cat "${PUBLISHER_REMOTE_BRANCH_FILE}" 2>/dev/null || true)" ]] || fail 'publisher changed remote after unexpected HEAD movement'
unset PUBLISHER_MOVE_HEAD_AFTER_DRY_RUN

diverged="$(new_repo diverged)"
base_sha="$(git -C "${diverged}" rev-parse HEAD)"
git -C "${diverged}" switch --quiet --create publisher-test
git -C "${diverged}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): refuse diverged target'
tree="$(git -C "${diverged}" rev-parse HEAD^{tree})"
remote_sha="$(git -C "${diverged}" commit-tree "${tree}" -p "${base_sha}" -m 'fix(test): divergent remote tip')"
export PUBLISHER_REMOTE_MAIN_SHA="${base_sha}"
export PUBLISHER_REMOTE_BRANCH_FILE="${test_root}/diverged-remote-sha"
export PUBLISHER_PUSH_LOG="${test_root}/diverged-push.log"
printf '%s\n' "${remote_sha}" >"${PUBLISHER_REMOTE_BRANCH_FILE}"
expect_rejected 'non-fast-forward target' bash -c "cd '${diverged}' && '${publisher}' publish"
[[ ! -s "${PUBLISHER_PUSH_LOG}" ]] || fail 'publisher attempted a push to a non-fast-forward target'

printf 'source publisher tests passed\n'

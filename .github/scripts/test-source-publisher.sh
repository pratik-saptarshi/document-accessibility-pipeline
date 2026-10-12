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
merge_output="$(cd "${merge}" && "${publisher}" check-remote refs/remotes/origin/main)"
[[ "${merge_output}" == *'GitHub-generated merge commit'* ]] || fail 'GitHub merge commit was not distinguished from a source commit'

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
    git_dir="$("${PUBLISHER_REAL_GIT}" rev-parse --git-dir)"
    printf '%s\t\tbranch %s of origin\n' "${sha}" "${ref}" >"${git_dir}/FETCH_HEAD"
    ;;
  push:--dry-run)
    printf 'dry-run\n' >>"${PUBLISHER_PUSH_LOG}"
    ;;
  push:origin)
    printf 'push\n' >>"${PUBLISHER_PUSH_LOG}"
    sha="$("${PUBLISHER_REAL_GIT}" rev-parse --verify HEAD)"
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
git -C "${push_repo}" -c user.name='CloudBot Neo-green' -c user.email='waynemanorsaptarshi@gmail.com' \
  commit --quiet --allow-empty -m 'fix(test): publish verified source commit'
export PUBLISHER_REAL_GIT="${real_git}"
export PUBLISHER_REMOTE_MAIN_SHA="${base_sha}"
export PUBLISHER_REMOTE_BRANCH_FILE="${test_root}/remote-branch-sha"
export PUBLISHER_PUSH_LOG="${test_root}/push.log"
export PATH="${fake_bin}:${PATH}"
(
  cd "${push_repo}"
  "${publisher}" publish
)
[[ "$(cat "${PUBLISHER_PUSH_LOG}")" == $'dry-run\npush' ]] || fail 'publisher did not dry-run before the ordinary push'
[[ "$(cat "${PUBLISHER_REMOTE_BRANCH_FILE}")" == "$(git -C "${push_repo}" rev-parse HEAD)" ]] || fail 'remote verification did not observe the pushed tip'

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

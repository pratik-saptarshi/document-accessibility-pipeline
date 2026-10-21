#!/usr/bin/env bash
set -euo pipefail

version=1.30.3
expected_fingerprint=13DD102B4DD69354D12DE5A83184863278B17FE7
expected_installer_sha256=b512bf24945503a33fc10738ecffc1f2dccaaf8593a448e94ee9ac86a9559a59
expected_cli_sha256=19f6edd5c92841a38d9b3b018a555b7ba173b13e4e0d29edecc232e5ba6c1d70
base_url=https://software.verapdf.org/releases/1.30
archive_name=verapdf-greenfield-${version}-installer.zip
if (( $# != 0 )); then
  echo "usage: provision-verapdf.sh" >&2
  exit 2
fi
workdir=$(mktemp -d)
install_root=$(mktemp -d "${RUNNER_TEMP:-${TMPDIR:-/tmp}}/verapdf-${version}.XXXXXX")
gnupg_home="$workdir/gnupg"
mkdir -m 700 "$gnupg_home"
export GNUPGHOME="$gnupg_home"
cleanup() { rm -rf -- "$workdir"; }
trap cleanup EXIT

curl --proto '=https' --fail --location --silent --show-error --max-time 180 \
  "$base_url/$archive_name" --output "$workdir/$archive_name"
curl --proto '=https' --fail --location --silent --show-error --max-time 30 \
  "$base_url/$archive_name.asc" --output "$workdir/$archive_name.asc"
curl --proto '=https' --fail --location --silent --show-error --max-time 30 \
  https://software.verapdf.org/keys/KEY --output "$workdir/KEY"

actual_fingerprint=$(gpg --batch --show-keys --with-colons "$workdir/KEY" \
  | awk -F: '$1 == "fpr" { print $10; exit }')
if [[ "$actual_fingerprint" != "$expected_fingerprint" ]]; then
  echo "veraPDF signing-key fingerprint mismatch" >&2
  exit 1
fi

printf '%s  %s\n' "$expected_installer_sha256" "$workdir/$archive_name" \
  | sha256sum --check --status
gpg --homedir "$gnupg_home" --batch --quiet --import "$workdir/KEY"
gpg --homedir "$gnupg_home" --batch --verify \
  "$workdir/$archive_name.asc" "$workdir/$archive_name"

unzip -q "$workdir/$archive_name" -d "$workdir"
cat > "$workdir/auto-install.xml" <<EOF
<AutomatedInstallation langpack="eng">
  <com.izforge.izpack.panels.htmlhello.HTMLHelloPanel id="welcome"/>
  <com.izforge.izpack.panels.target.TargetPanel id="install_dir">
    <installpath>$install_root</installpath>
  </com.izforge.izpack.panels.target.TargetPanel>
  <com.izforge.izpack.panels.packs.PacksPanel id="sdk_pack_select">
    <pack index="0" name="veraPDF GUI" selected="false"/>
    <pack index="1" name="veraPDF Mac and *nix Scripts" selected="true"/>
    <pack index="2" name="veraPDF Validation model" selected="true"/>
    <pack index="3" name="veraPDF Documentation" selected="false"/>
    <pack index="4" name="veraPDF Sample Plugins" selected="false"/>
  </com.izforge.izpack.panels.packs.PacksPanel>
  <com.izforge.izpack.panels.install.InstallPanel id="install"/>
  <com.izforge.izpack.panels.finish.FinishPanel id="finish"/>
</AutomatedInstallation>
EOF

installer_dir="$workdir/verapdf-greenfield-$version"
java -jar "$installer_dir/verapdf-izpack-installer-$version.jar" \
  "$workdir/auto-install.xml"
cli_jar="$install_root/bin/cli-$version.jar"
printf '%s  %s\n' "$expected_cli_sha256" "$cli_jar" | sha256sum --check --status
version_output=$(java -jar "$cli_jar" --version)
if ! grep -Fq "veraPDF $version" <<<"$version_output"; then
  echo "installed veraPDF version mismatch" >&2
  exit 1
fi
if [[ -n "${GITHUB_ENV:-}" ]]; then
  {
    printf 'PDFUA_REAL_CLI_JAR=%s\n' "$cli_jar"
    printf 'PDFUA_REAL_CLI_SHA256=%s\n' "$expected_cli_sha256"
  } >> "$GITHUB_ENV"
fi
printf 'Verified and installed veraPDF %s; installer SHA-256 %s; CLI JAR SHA-256 %s\n' \
  "$version" "$expected_installer_sha256" "$expected_cli_sha256"

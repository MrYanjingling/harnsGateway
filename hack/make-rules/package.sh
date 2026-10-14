#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

HARNSGATEWAY_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
source "${HARNSGATEWAY_ROOT}/hack/lib/init.sh"

# Package name and the module directories laid out inside the archive.
PACKAGE_NAME="${PACKAGE_NAME:-harns-platform}"
PACKAGE_MODULES=("web" "gateway" "platform" "eam" "ems")

# Build the binaries first, reusing the standard build rules.
"${HARNSGATEWAY_ROOT}/hack/make-rules/build.sh" "$@"

# Resolve the platform that was built (defaults to the host platform).
platforms=()
IFS=" " read -ra platforms <<< "${HARNSGATEWAY_BUILD_PLATFORMS:-}"
target_platform="${platforms[0]:-$(harnsgateway::golang::host_platform)}"

host_os="${target_platform%%/*}"
host_arch="${target_platform##*/}"
bin_dir="${HARNSGATEWAY_OUTPUT_BINPATH}/${target_platform}"

release_dir="${HARNSGATEWAY_OUTPUT}/release"
staging_dir="${release_dir}/${PACKAGE_NAME}"

echo "Packaging ${PACKAGE_NAME} for ${target_platform}"

rm -rf "${staging_dir}"
mkdir -p "${staging_dir}"

for module in "${PACKAGE_MODULES[@]}"; do
  mkdir -p "${staging_dir}/${module}"
done

# "web" holds the compiled frontend assets and carries no bin/etc/log layout.
for module in "${PACKAGE_MODULES[@]}"; do
  if [[ "${module}" == "web" ]]; then
    continue
  fi

  mkdir -p "${staging_dir}/${module}/bin" "${staging_dir}/${module}/etc" "${staging_dir}/${module}/log"

  if [[ -f "${bin_dir}/${module}" ]]; then
    cp -p "${bin_dir}/${module}" "${staging_dir}/${module}/bin/"
  else
    echo "Warning: binary '${module}' not found in ${bin_dir}"
  fi

  if [[ -d "${HARNSGATEWAY_ROOT}/etc/${module}" ]]; then
    cp -Rp "${HARNSGATEWAY_ROOT}/etc/${module}/." "${staging_dir}/${module}/etc/"
  fi
done

# Determine a version suffix, falling back to the short commit when there is no tag.
harnsgateway::version::get_version_vars
version="${HARNSGATEWAY_GIT_VERSION:-}"
if [[ -z "${version}" ]]; then
  version="$(git --work-tree "${HARNSGATEWAY_ROOT}" rev-parse --short HEAD 2>/dev/null || echo dev)"
fi

archive_name="${PACKAGE_NAME}-${version}-${host_os}_${host_arch}.zip"
archive="${release_dir}/${archive_name}"

rm -f "${archive}"
(
  cd "${release_dir}"
  zip -qr "${archive_name}" "${PACKAGE_NAME}"
)

echo "Packaged: ${archive}"

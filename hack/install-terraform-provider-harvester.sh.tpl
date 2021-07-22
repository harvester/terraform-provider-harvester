#!/usr/bin/env bash
[[ -n $DEBUG ]] && set -x
set -eou pipefail

usage() {
    cat <<HELP
USAGE:
    install-terraform-provider-harvester.sh
HELP
}

version=%VERSION%
arch=linux_%ARCH%
terraform_harvester_provider_bin=./terraform-provider-harvester

terraform_harvester_provider_dir="${HOME}/.terraform.d/plugins/registry.terraform.io/harvester/harvester/${version}/${arch}/"
mkdir -p "${terraform_harvester_provider_dir}"
cp ${terraform_harvester_provider_bin} "${terraform_harvester_provider_dir}/terraform-provider-harvester_v${version}"

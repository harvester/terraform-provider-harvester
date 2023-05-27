#!/usr/bin/env bash
[[ -n $DEBUG ]] && set -x
set -eou pipefail

usage() {
    cat <<HELP
USAGE:
    setup-test-env.sh [KUBE_CONFIG_FILE]
HELP
}

if [ $# -lt 1 ]; then
    usage
    exit 1
fi

KUBE_CONFIG_FILE=$1

# build dev image
REPO="" TAG=dev make

# create tf container
docker rm -f tf
docker run -itd --name tf rancher/terraform-provider-harvester:dev bash

# copy kubeconfig to the tf container
docker cp "${KUBE_CONFIG_FILE}" tf:/data/kubeconfig

# exec to the tf container
docker exec -it tf bash
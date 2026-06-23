ARG MK_GOLANGCI_LINT_IMAGE
ARG MK_PACKAGE_BASE registry.suse.com/bci/bci-base:16.0
FROM ${MK_GOLANGCI_LINT_IMAGE} AS golangci-lint

FROM golang:1.25-bookworm AS buildenv
ARG TERRAFORM_VERSION
ARG TERRAFORM_SUM_amd64
ARG TERRAFORM_SUM_arm64
ARG TARGETPLATFORM
ENV GOTOOLCHAIN=auto

RUN --mount=type=cache,target=/var/lib/apt/lists apt-get update -qq \
 && apt-get install -y --no-install-recommends \
  unzip

ENV ARCH=${TARGETPLATFORM#linux/}
RUN curl -sfL -o /tmp/terraform.zip "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${ARCH}.zip" \
 && TERRAFORM_SUM=$([ "${ARCH}" = "amd64" ] && echo "${TERRAFORM_SUM_amd64}" || echo "${TERRAFORM_SUM_arm64}") \
 && echo "${TERRAFORM_SUM}  /tmp/terraform.zip" | sha256sum -c - \
 && unzip /tmp/terraform.zip -d /tmp/ && mv /tmp/terraform /usr/bin/ && rm -f /tmp/terraform.zip \
 && terraform version
COPY --from=golangci-lint /usr/bin/golangci-lint /usr/local/bin/golangci-lint

# ---- base ----
FROM buildenv AS base
ARG MK_REPO
ARG MK_REPO_ID
WORKDIR /go/src/${MK_REPO}
# to exclude some files, add them in .dockerignore
COPY . .

# ---- build ----
FROM base AS build
ARG MK_REPO
ARG MK_REPO_ID
RUN --mount=type=cache,target=/go/pkg/mod,id=harvester-go-mod-${MK_REPO_ID} \
    --mount=type=cache,target=/go/src/${MK_REPO}/.cache/go-build,id=harvester-go-build-${MK_REPO_ID} \
    <<EOF
#!/bin/bash -e

go generate

mkdir -p bin
[ "$(uname)" != "Darwin" ] && LINKFLAGS="-extldflags -static -s"
CGO_ENABLED=0 GOARCH=amd64 go build -ldflags "-X main.VERSION=$VERSION $LINKFLAGS" -o bin/terraform-provider-harvester-amd64
CGO_ENABLED=0 GOARCH=arm64 go build -ldflags "-X main.VERSION=$VERSION $LINKFLAGS" -o bin/terraform-provider-harvester-arm64
EOF

# ---- test ----
FROM base AS test
ARG MK_REPO
ARG MK_REPO_ID
RUN --mount=type=cache,target=/go/pkg/mod,id=harvester-go-mod-${MK_REPO_ID} \
    --mount=type=cache,target=/go/src/${MK_REPO}/.cache/go-build,id=harvester-go-build-${MK_REPO_ID} \
    --mount=type=secret,id=codecov_token_${MK_REPO_ID},env=CODECOV_TOKEN \
    <<EOF
#!/bin/bash -e

if [ -f ./kubeconfig_test.yaml ] ; then
  export KUBECONFIG="$(pwd)/kubeconfig_test.yaml"
  export TF_ACC=1
  # Avoid timeout after 10 minutes https://pkg.go.dev/cmd/go#hdr-Testing_flags
  export EXTRA_OPTIONS=("-timeout" "0")
fi

echo Running tests:
go test \
  -v \
  -cover \
  -coverprofile=coverage.out \
  -tags=test \
  "${EXTRA_OPTIONS[@]}" \
  . ./...

go tool cover \
  -html=coverage.out \
  -o coverage.html
EOF

# ---- validate ----
FROM base AS validate
ARG MK_REPO
ARG MK_REPO_ID
ARG PROVIDER_VERSION v0.0.0-dev
RUN --mount=type=cache,target=/go/pkg/mod,id=harvester-go-mod-${MK_REPO_ID} \
    --mount=type=cache,target=/go/src/${MK_REPO}/.cache/go-build,id=harvester-go-build-${MK_REPO_ID} \
    <<EOF
#!/bin/bash -e

echo Running validation

PACKAGES="$(go list ./...)"

echo Running validation: golangci-lint
golangci-lint run --timeout 5m

echo Running validation: go fmt
test -z "$(go fmt ${PACKAGES} | tee /dev/stderr)"

echo "Running dirty check"

go generate

if echo "$PROVIDER_VERSION" | grep dirty ; then
    echo "Git is dirty"
    git status
    git diff
    exit 1
fi

echo "All clean"
EOF

# ---- build output ----
FROM scratch AS build-output
ARG MK_REPO
COPY --from=build /go/src/${MK_REPO}/bin/ /bin/
COPY --from=build /go/src/${MK_REPO}/docs/ /docs/

# ---- package output ----
FROM ${MK_PACKAGE_BASE} as package
ARG PROVIDER_VERSION v0.0.0-dev
ARG MK_REPO
ARG TARGETARCH

ENV ARCH=${TARGETPLATFORM#linux/}
ENV PROVIDERS_DIR /root/.terraform.d/plugins/terraform.local/local/harvester
ENV PROVIDER_DIR ${PROVIDERS_DIR}/${PROVIDER_VERSION}/linux_${ARCH}
# hadolint ignore=DL3037
RUN zypper -n rm container-suseconnect && \
    zypper -n install unzip curl vim && \
    zypper -n clean -a && rm -rf /tmp/* /var/tmp/* /usr/share/doc/packages/*

RUN mkdir -p /data ${PROVIDER_DIR}
COPY --from=build \
  /go/src/${MK_REPO}/bin/terraform-provider-harvester-${TARGETARCH} \
  /${PROVIDER_DIR}/terraform-provider-harvester_v${PROVIDER_VERSION}

COPY --from=buildenv /usr/bin/terraform /usr/bin/terraform
RUN cat <<EOF > /data/provider.tf
terraform {
  required_providers {
    harvester = {
      source = "terraform.local/local/harvester"
      version = "${PROVIDER_VERSION}"
    }
  }
}
provider "harvester" {
  kubeconfig = "kubeconfig"
EOF

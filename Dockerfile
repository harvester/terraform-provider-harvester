ARG MK_GOLANGCI_LINT_IMAGE=golangci/golangci-lint:v2.12.2-alpine@sha256:91b27804074a0bacea298707f016911e60cf0cdbc6c7bf5ccacb5f0606d18d60
ARG MK_PACKAGE_BASE=registry.suse.com/bci/bci-base:16.1
ARG MK_REPO=github.com/harvester/terraform-provider-harvester
ARG MK_REPO_ID=default
ARG PROVIDER_VERSION=0.0.0-dev
ARG TERRAFORM_VERSION=1.4.6
ARG TERRAFORM_SUM_amd64=e079db1a8945e39b1f8ba4e513946b3ab9f32bd5a2bdf19b9b186d22c5a3d53b
ARG TERRAFORM_SUM_arm64=b38f5db944ac4942f11ceea465a91e365b0636febd9998c110fbbe95d61c3b26
FROM ${MK_GOLANGCI_LINT_IMAGE} AS golangci-lint

FROM registry.suse.com/bci/golang:1.26 AS buildenv
ARG TERRAFORM_VERSION
ARG TERRAFORM_SUM_amd64
ARG TERRAFORM_SUM_arm64
ARG TARGETPLATFORM
ENV GOTOOLCHAIN=auto

RUN zypper -n rm container-suseconnect \
 && zypper -n install git curl unzip \
 && zypper -n clean -a \
 && rm -rf /tmp/* /var/tmp/* /usr/share/doc/packages/*

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
# Create a temporary Git baseline inside the image so the dirty check only
# reports files changed by validation commands.
RUN rm -rf .git \
 && git config --global user.email "ci@example.com" \
 && git config --global user.name "ci" \
 && git init 2>/dev/null \
 && echo ".cache/" >> .git/info/exclude \
 && git add . \
 && git commit -q -m "commit for validate"
RUN --mount=type=cache,target=/go/pkg/mod,id=harvester-go-mod-${MK_REPO_ID} \
    --mount=type=cache,target=/go/src/${MK_REPO}/.cache/go-build,id=harvester-go-build-${MK_REPO_ID} \
    <<EOF
#!/bin/bash -e

echo Running validation

PACKAGES="$(go list ./...)"

echo Running validation: golangci-lint
golangci-lint run --timeout 5m

echo Running validation: go fmt
go fmt ${PACKAGES}

echo "Running dirty check"

go generate

if [ -n "$(git status --porcelain | tee /dev/stderr)" ]; then
    echo "Git is dirty"
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
FROM ${MK_PACKAGE_BASE} AS package
ARG PROVIDER_VERSION
ARG MK_REPO
ARG TARGETARCH

ENV ARCH=${TARGETARCH}
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
}
EOF

WORKDIR /data

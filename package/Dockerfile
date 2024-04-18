# syntax=docker/dockerfile:1.7.0

FROM registry.suse.com/bci/bci-base:15.5

ARG TARGETPLATFORM
RUN if [ "$TARGETPLATFORM" != "linux/amd64" ] && [ "$TARGETPLATFORM" != "linux/arm64" ]; then \
    echo "Error: Unsupported TARGETPLATFORM: $TARGETPLATFORM" && \
    exit 1; \
    fi

# hadolint ignore=DL3037
RUN zypper -n rm container-suseconnect && \
    zypper -n install unzip curl vim && \
    zypper -n clean -a && rm -rf /tmp/* /var/tmp/* /usr/share/doc/packages/*

ENV ARCH ${TARGETPLATFORM#linux/}
ENV KERNEL_ARCH linux_${ARCH}
# install terraform
ENV TERRAFORM_VERSION 1.4.6
ENV TERRAFORM_URL https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_${KERNEL_ARCH}.zip

RUN curl -sfL -o terraform.zip ${TERRAFORM_URL} && \
    unzip terraform.zip -d /tmp/ && mv /tmp/terraform /usr/bin/ && rm -f terraform.zip && \
    terraform version

# place provider into plugins directory
ARG PROVIDER_VERSION=0.0.0-dev
ENV PROVIDERS_DIR /root/.terraform.d/plugins/terraform.local/local/harvester
ENV PROVIDER_DIR ${PROVIDERS_DIR}/${PROVIDER_VERSION}/${KERNEL_ARCH}
RUN mkdir -p ${PROVIDER_DIR}
COPY ./terraform-provider-harvester-${ARCH} ${PROVIDER_DIR}/terraform-provider-harvester_v${PROVIDER_VERSION}
RUN mkdir -p /data
COPY ./provider.tf /data/provider.tf
WORKDIR /data

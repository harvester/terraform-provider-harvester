FROM registry.suse.com/bci/bci-base:15.3

RUN zypper -n rm container-suseconnect && \
    zypper -n install unzip=6.00 curl=7.66.0 vim=8.2.5038 && \
    zypper -n clean -a && rm -rf /tmp/* /var/tmp/* /usr/share/doc/packages/*

ARG ARCH=amd64
ENV KERNEL_ARCH linux_${ARCH}
# install terraform
ENV TERRAFORM_VERSION 1.1.5
ENV TERRAFORM_URL https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_${KERNEL_ARCH}.zip

RUN curl -sfL -o terraform.zip ${TERRAFORM_URL} && \
    unzip terraform.zip -d /tmp/ && mv /tmp/terraform /usr/bin/ && rm -f terraform.zip && \
    terraform version

# place provider into plugins directory
ARG PROVIDER_VERSION=0.0.0-master
ENV PROVIDERS_DIR /root/.terraform.d/plugins/registry.terraform.io/harvester/harvester
ENV PROVIDER_DIR ${PROVIDERS_DIR}/${PROVIDER_VERSION}/${KERNEL_ARCH}
RUN mkdir -p ${PROVIDER_DIR}
COPY ./terraform-provider-harvester ${PROVIDER_DIR}/terraform-provider-harvester_v${PROVIDER_VERSION}

RUN mkdir -p /data
WORKDIR /data
ROOT := $(realpath $(dir $(realpath $(firstword $(MAKEFILE_LIST)))))
DOCKER_BUILDKIT := 1
export DOCKER_BUILDKIT

ifdef CI
	BOLD  :=
	CYAN  :=
	RESET :=
else
	BOLD  := \033[1m
	CYAN  := \033[36m
	RESET := \033[0m
endif
BANNER = @printf "$(BOLD)$(CYAN)[target: $@]$(RESET)\n"

MK_SYSTEM_ID := $(strip $(shell \
		if [ -s /etc/machine-id ]; then \
				cat /etc/machine-id 2>/dev/null; \
		elif command -v hostname >/dev/null 2>&1; then \
				hostname 2>/dev/null; \
		else \
				echo -n "unknown"; \
		fi))

MK_REPO_ID          := $(shell printf '%s' "$(ROOT)$(MK_SYSTEM_ID)" | sha256sum | cut -c1-8)
MK_IMAGE_TAG        := $(shell git describe --tags --always --dirty)
MK_CODECOV_TOKEN    ?=
MK_DOCKER_PROGRESS  ?= plain

MK_CODECOV_SECRET_ARG  := --secret id=codecov_token_$(MK_REPO_ID),env=MK_CODECOV_TOKEN --no-cache-filter=test
DOCKER_BUILD := \
	docker build \
		--progress=$(MK_DOCKER_PROGRESS) \
		--build-arg MK_REPO_ID=$(MK_REPO_ID) \
		-f $(ROOT)/Dockerfile $(ROOT)

.PHONY: ci version validate build test package

# ---- Directories ----
$(ROOT)/bin:
	@mkdir -p $@

$(ROOT)/docs:
	@mkdir -p $@

# ---- Print Docker image tag version ----
version:
	@printf '%s\n' '$(MK_IMAGE_TAG)'

# ---- Validate static checks and generated files ----
validate:
	$(BANNER)
	$(DOCKER_BUILD) --target validate

# ---- Compile harvester-terraform-provider binaries ----
build: $(ROOT)/bin $(ROOT)/docs
	$(BANNER)
	$(DOCKER_BUILD) --target build-output --output type=local,dest=.

# ---- Test ----
test: validate
	$(BANNER)
	$(DOCKER_BUILD) $(if $(MK_CODECOV_TOKEN),$(MK_CODECOV_SECRET_ARG)) --target test

# ---- Package harvester-terraform-provider image ----
package: build
	$(BANNER)
	$(DOCKER_BUILD) --target package -t terraform-provider-harvester:$(MK_IMAGE_TAG)

ci: validate build test package
	$(BANNER)

.DEFAULT_GOAL := default
default: build test package

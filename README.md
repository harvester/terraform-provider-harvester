Terraform Provider for Harvester
==================================

Requirements
------------

- [Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
- [Go](https://golang.org/doc/install) 1.16 to build the provider plugin

Building The Provider
---------------------

Clone repository

```bash
$ git clone git@github.com:harvester/terraform-provider-harvester
```

Enter the provider directory and build the provider

This will build the provider and put the provider binary in `./bin`.

```bash
$ cd terraform-provider-harvester
$ make
```

Install the provider
----------------------
The expected location for the Harvester provider for that target platform within one of the local search directories would be the following:
```bash
registry.terraform.io/harvester/harvester/0.1.0/linux_amd64/terraform-provider-harvester_v0.1.0
```

The default location for locally-installed providers is one of the following, depending on which operating system you are running Terraform under:
* Windows: %APPDATA%\terraform.d\plugins
* All other systems: ~/.terraform.d/plugins

Place the provider into the plugins directory, for example:
```bash
version=0.1.0
arch=linux_amd64
terraform_harvester_provider_dir=~/.terraform.d/plugins/registry.terraform.io/harvester/harvester/${version}/${arch}/
mkdir -p ${terraform_harvester_provider_dir}
cp ./bin/terraform-provider-harvester ${terraform_harvester_provider_dir}/terraform-provider-harvester_v${version}
```

Using the provider
----------------------
After placing it into your plugins directory,  run `terraform init` to initialize it.
Documentation about the provider specific configuration options can be found on the [docs directory](docs).
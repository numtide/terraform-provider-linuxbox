# Linuxbox terraform provider

Basic building block for Seed DevOps.

This provider allows:
* Executing of commands via SSH as a resource.
* Calculating checksum from list of source paths/files.
* Building Docker containers from Dockerfiles.
* Copying built Docker images to destinations hosts vis SSH.
* Controlling Docker containers on destination hosts via SSH.
* Creating Docker networks on destination hosts via SSH.

# Installation

Easiest and most efficient way of installing the provider is to generate the provider shim using [generate-terraform-provider-shim](https://github.com/numtide/generate-terraform-provider-shim):

```sh
$ generate-terraform-provider-shim numtide/terraform-provider-linuxbox
```

Generated provider shims (one per found ARCH of the provider) are a small Bash script and can be easily checked in with the rest of the terraform files.

if a version that satisfies Semver constraints is required, this can be specified at generation time:

```sh
$ generate-terraform-provider-shim --version '< 0.2.0, >= 0.1.0' numtide/terraform-provider-linuxbox
```


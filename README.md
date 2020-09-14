# Linuxbox terraform provider

Basic building block for Seed DevOps.

This provider allows:
* Executing of commands via SSH as a resource.
* Calculating checksum from list of source paths/files.
* Building Docker containers from Dockerfiles.
* Copying built Docker images to destinations hosts vis SSH.
* Controlling Docker containers on destination hosts via SSH.
* Creating Docker networks on destination hosts via SSH.

## Installation

Easiest and most efficient way of installing the provider is to generate the provider shim using [generate-terraform-provider-shim](https://github.com/numtide/generate-terraform-provider-shim):

```console
$ generate-terraform-provider-shim numtide/terraform-provider-linuxbox
```

Generated provider shims (one per found ARCH of the provider) are a small Bash script and can be easily checked in with the rest of the terraform files.

if a version that satisfies Semver constraints is required, this can be specified at generation time:

```console
$ generate-terraform-provider-shim --version '< 0.2.0, >= 0.1.0' numtide/terraform-provider-linuxbox
```

## Use

### Configuring Provider

Provider accepts one optional argument: `ssh_session_limit`.
This is the limits number of sessions that will be open through SSH connection to a host.
Current default limit is `5`.

Sample provider declaration with setting the `ssh_session_limit` lower looks like this:

```hcl
provider "linuxbox" {
  ssh_session_limit = 3
}
```

### SSH Configuration used by every SSH resource.

Every Linuxbox resource that uses SSH will accept following parameters:

* **ssh_key**: This is the private key used to authenticate user when connecting to the destination host.

* **ssh_user**: Username used to authenticated when connecting to the destination host.
By default, this username is `root`.
If the username is not root, make sure that the user has the right permissions on the destination host to execute required operations.

* **host_address**: Address (dns name or IP address) of the target host.

### Performing setup of a remote machine using SSH.
Philosophy of Linuxbox is similar to the one of Ansible.
We don't require any kind of agent or a service to be run on the remote machine apart from SSH.
Every step of a machine setup can be represented as a separate Terraform resource.
By doing so, we make sure that setup steps are executed only once and in order given
by `depends_on` or other dependency resolving mechanism of Terraform.
This enables parallelisation of execution of certain tasks (for example: adding a swap and installing Docker) which this will be automatically handled by Terraform.

Every setup step is defined using `linuxbox_run_setup` resource.

Since setup is a Terraform resource, user has to provide 3 parts to it:
* List of commands to be executed to perform setup (`setup`). This performed when `terraform apply` is executed.
* Command that will tell if the result of the setup is available on the machine (`check`). that is what `terraform plan` will query)
* Command that will remove the result of the setup ... for examle, removing the installed package or removing the swap.

only `setup` is mandatory.
If `check` is omitted, plan will alway report resource being present (can be misleading if in the meantime someone has logged in into the machine and has deleted the installed package).

If `delete` is omitted, removing/destroying the resource in terrafom won't have any effect on what is installed on the destination machine.

For example, following setup will install docker on the target ubuntu 18.04 machine:

```hcl
resource "linuxbox_run_setup" "install_docker" {
  host_address = digitalocean_droplet.test.ipv4_address
  ssh_key      = tls_private_key.ssh_key.private_key_pem

  setup = [
    "apt update",
    "apt install -y apt-transport-https ca-certificates curl software-properties-common",
    "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -",
    "add-apt-repository \"deb [arch=amd64] https://download.docker.com/linux/ubuntu bionic stable\"",
    "apt update",
    "apt install -y docker-ce",
  ]

  check = "docker -v"

  delete = "apt-get purge -y docker-ce docker-ce-cli"

}
```

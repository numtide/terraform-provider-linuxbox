# `linuxbox_run_setup` Resource

Executes setup scripts on the target host. This is mainly useful to configure
machines after startup.

## Example Usage

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

## Argument Reference

* `host_address` - (Required) Machine hostname to connect to.
* `ssh_key`      - (Required) Machine SSH key to connect with.
* `ssh_user`     - (Optional) Machine SSH user to connect with (default: "root").

* `setup`        - (Required) A list of commands to run.
* `check`        - (Optional) Verify if the setup needs to run.
* `delete`       - (Optional) Run on deletion.

## Attribute Reference

None

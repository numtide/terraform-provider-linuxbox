# `linuxbox_network` Resource

Declares a Docker network on the target host.

## Example Usage

```hcl
resource "linuxbox_network" "my_network" {
  host_address = digitalocean_droplet.test.ipv4_address
  ssh_key      = tls_private_key.ssh_key.private_key_pem

  name = "my_network"
}
```

## Argument Reference

* `host_address` - (Required) Machine hostname to connect to.
* `ssh_key`      - (Required) Machine SSH key to connect with.
* `ssh_user`     - (Optional) Machine SSH user to connect with (default: "root").

* `name`         - (Required) Name of the docker network to create.

## Attribute Reference

None

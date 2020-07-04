# `linuxbox_ssh_authorized_key` Resource

Adds a SSH public key to the `.ssh/authorized_keys` file on the target host.

## Example Usage

```hcl
provider "linuxbox" {}

resource "linuxbox_ssh_authorized_key" "my_instance" {
  host_address = digitalocean_droplet.test.ipv4_address
  ssh_key      = tls_private_key.ssh_key.private_key_pem
  key_to_add   = <<EOM
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDGB1Pog97SWdV2UEA40V+3bML+lSZXEd48zCRlS/eGbY3rsXfgUXb5FIBulN9cET9g0OOAKeCZBR1Y2xXofiHDYkhk298rHDuir6cINuoMGUO7VsygUfKguBy63QMPHYnJBE1h+6sQGu/3X9G2o/0Ys2J+lZv4+N7Hqolhbg/Cu6/LUCsJM/udqTVwJGEqszDWPtuuTAIS6utB1QdL9EZT5WBb1nsNyHnIlCnoDKZvrrO9kM0FGKhjJG2skd3+NqmLhYIDhRhZvRnL9c8U8uozjbtj/N8L/2VCRzgzKmvu0Y1cZMWeAAdyqG6LoyE7xGO+SF4Vz1x6JjS9VxnZipIB zimbatm@nixos
  EOM
}
```

## Argument Reference

* `host_address` - (Required) Machine hostname to connect to.
* `ssh_key`      - (Required) Machine SSH key to connect with.
* `ssh_user`     - (Optional) Machine SSH user to connect with (default: "root").

* `key_to_add`   - (Required) SSH public key to add to the machine.

## Attribute Reference

None

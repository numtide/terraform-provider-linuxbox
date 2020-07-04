# linuxbox_swap Resources

Creates and mounts a `/swapfile` in the target machine.

## Example Usage

```hcl
provider "linuxbox" {}

resource "linuxbox_swap" "my_instance" {
  host_address = digitalocean_droplet.test.ipv4_address
  ssh_key      = tls_private_key.ssh_key.private_key_pem
  swap_size    = 10 * 1024 * 1024 # 10 GB
}
```

## Argument Reference

* `host_address` - (Required) Machine hostname to connect to.
* `ssh_key`      - (Required) Machine SSH key to connect to.
* `swap_size`    - (Required) Size of the swap, in bytes.

## Attribute Reference

None

# linuxbox_swap Resources

Creates and mounts a `/swapfile` in the target machine.

## Example Usage

```hcl
provider "linuxbox" {}

resource "linuxbox_swap" "my_instance" {
  ssh_key = TODO
  host_address = TODO
  swap_size = 10 * 1024 * 1024 # 10 GB
}
```

## Argument Reference

* `ssh_key` - (Required) Machine SSH key to connect to.
* `host_address` - (Required) Machine hostname to connect to.
* `swap_size` - (Required) Size of the swap, in bytes.

## Attribute Reference

None

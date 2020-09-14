# `linuxbox_text_file` Resource

Creates a file on the target host if missing.

## Example Usage

```hcl
resource "linuxbox_text_file" "authorized_keys" {
  host_address = digitalocean_droplet.test.ipv4_address
  ssh_key      = tls_private_key.ssh_key.private_key_pem

  path  = "/root/.ssh/authorized_keys"
  content = <<CONTENT
  # ...
  CONTENT
  owner = 0
  group = 0
  mode  = 600
}
```

## Argument Reference

* `host_address` - (Required) Machine hostname to connect to.
* `ssh_key`      - (Required) Machine SSH key to connect with.
* `ssh_user`     - (Optional) Machine SSH user to connect with (default: "root").

* `path`         - (Required) Path of the file to create.
* `content`      - (Required) Content of the file to create.
* `owner`        - (Optional) User ID of the folder (default: 0).
* `group`        - (Optional) Group ID of the folder (default: 0).
* `mode`         - (Optional) File mode (default: 644).

## Attribute Reference

None

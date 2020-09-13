# `directory` Resource

Creates a directory on the target host if missing.

## Example Usage

```hcl
resource "linuxbox_directory" "secrets_folder" {
  host_address = digitalocean_droplet.test.ipv4_address
  ssh_key      = tls_private_key.ssh_key.private_key_pem

  path  = "/etc/secrets"
  owner = 0
  group = 0
  mode  = 700
}
```

## Argument Reference

* `host_address` - (Required) Machine hostname to connect to.
* `ssh_key`      - (Required) Machine SSH key to connect with.
* `ssh_user`     - (Optional) Machine SSH user to connect with (default: "root").

* `path`         - (Required) Path of the folder to create.
* `owner`        - (Optional) User ID of the folder (default: 0).
* `group`        - (Optional) Group ID of the folder (default: 0).
* `mode`         - (Optional) Folder mode (default: 755).

## Attribute Reference

None

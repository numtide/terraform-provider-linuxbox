# `linuxbox_text_file` Resource

Reads a file from the target host.

## Example Usage

```hcl
datasource "linuxbox_text_file" "authorized_keys" {
  host_address = digitalocean_droplet.test.ipv4_address
  ssh_key      = tls_private_key.ssh_key.private_key_pem

  path  = "/root/.ssh/authorized_keys"
}
```

## Argument Reference

* `host_address` - (Required) Machine hostname to connect to.
* `ssh_key`      - (Required) Machine SSH key to connect with.
* `ssh_user`     - (Optional) Machine SSH user to connect with (default: "root").
* `path`         - (Required) Path of the file to create.

## Attribute Reference

- `content` - Content of the file

None

# linuxbox_source_hash Data Source

Read data from the give source paths and calculate a unique hash from it.
Used to detect when sources change.

## Example Usage

```hcl
provider "linuxbox" {}

data "linuxbox_source_hash" "data" {
  sources = [
    "${path.module}/src"
  ]
}

output "source_hash" {
  value = data.linuxbox_source_hash.data.hash
}
```

## Argument Reference

* `sources` - (Required) A list of paths to hash

## Attribute Reference

* `hash` - The calculated hash from all the sources.

# `linuxbox_docker_build` Resource

This resource can be used to build docker images on the host that runs
Terraform. Depends on `docker` to be installed.

## Example Usage

```hcl
data "linuxbox_source_hash" "my_image" {
  sources = [
    "${path.module}/Dockerfile",
    "${path.module}/image",
  ]
}

resource "linuxbox_docker_build" "my_image" {
  source_dir  = "${path.module}/image"
  source_hash = data.linuxbox_source_hash.my_image.hash
  dockerfile  = "${path.module}/Dockerfile"
}
```

## Argument Reference

* `source_dir` - (Required) Folder path to build.
* `source_hash` - (Required) Hash of the source.
* `dockerfile` - (Optional) Defaults to `./Dockerfile`.

## Attribute Reference

* `image_id` - ID of the image that has been built.

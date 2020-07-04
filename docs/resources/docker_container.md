# `linuxbox_docker_container` Resource

Executes a Docker image on the target host in the background.

## Example Usage

TODO

## Argument Reference

* `host_address` - (Required) Machine hostname to connect to.
* `ssh_key`      - (Required) Machine SSH key to connect with.
* `ssh_user`     - (Optional) Machine SSH user to connect with (default: "root").

* `image_id`     - (Required) Name of the docker image to run.
* `ports`        - (Optional) List of ports to bind to.
* `caps`         - (Optional) List of strings.
* `volumes`      - (Optional) List of strings.
* `labels`       - (Optional) List of strings.
* `env`          - (Optional) List of strings.
* `privileged`   - (Optional) Defaults to false.
* `network`      - (Optional) String.
* `args`         - (Optional) List of arguments to run.
* `restart`      - (Optional) String.
* `name`         - (Optional) Name of the docker container to run.

## Attribute Reference

* `container_id` - (Optional) ID of the running container.

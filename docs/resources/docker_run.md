# `linuxbox_docker_run` Resource

Executes a Docker image on the target host once.

## Example Usage

TODO

## Argument Reference

* `host_address` - (Required) Machine hostname to connect to.
* `ssh_key`      - (Required) Machine SSH key to connect to.
* `ssh_user`     - (Optional) Machine SSH user.

* `image_id`     - (Required) Name of the docker image to run.
* `ports`        - (Optional) List of ports to bind to.
* `caps`         - (Optional) List of strings.
* `volumes`      - (Optional) List of strings.
* `labels`       - (Optional) List of strings.
* `env`          - (Optional) List of strings.
* `privileged`   - (Optional) Defaults to false.
* `network`      - (Optional) String.
* `args`         - (Optional) List of arguments to run.

## Attribute Reference

* `stdout` - Output of the command.
* `stderr` - Output of the command.

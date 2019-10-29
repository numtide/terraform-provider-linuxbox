provider "digitalocean" {
}

provider "linuxbox" {
}

resource "digitalocean_droplet" "test" {
    image  = "ubuntu-18-04-x64"
    name   = "terraform-test-1"
    region = "lon1"
    size   = "s-1vcpu-1gb"
    ssh_keys = [digitalocean_ssh_key.terraform.fingerprint, data.digitalocean_ssh_key.terraform.fingerprint]
}

resource "tls_private_key" "ssh_key" {
    algorithm = "ECDSA"
    ecdsa_curve = "P521"
}


resource "digitalocean_ssh_key" "terraform" {
    name       = "Terraform Test"
    public_key = tls_private_key.ssh_key.public_key_openssh
}

data "digitalocean_ssh_key" "terraform" {
    name = "my-laptop"
}

resource "linuxbox_swap" "host_swap" {
    host_address = digitalocean_droplet.test.ipv4_address
    ssh_key = tls_private_key.ssh_key.private_key_pem
    swap_size = "100m"
}

resource "linuxbox_docker" "docker" {
    host_address = digitalocean_droplet.test.ipv4_address
    ssh_key = tls_private_key.ssh_key.private_key_pem
}

resource "linuxbox_docker_copy_image" "service" {
    depends_on = [linuxbox_docker.docker]
    host_address = digitalocean_droplet.test.ipv4_address
    ssh_key = tls_private_key.ssh_key.private_key_pem
    image_id = linuxbox_docker_build.sample_service.image_id
}

data "linuxbox_source_hash" "sample_service" {
    source_dirs = ["sample_service"]
}

resource "linuxbox_docker_build" "sample_service" {
    source_dir = "${path.module}/sample_service"
    source_hash = data.linuxbox_source_hash.sample_service.hash
}
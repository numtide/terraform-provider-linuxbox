provider "digitalocean" {
}

provider "linuxbox" {
}

resource "digitalocean_droplet" "test" {
    image  = "ubuntu-18-04-x64"
    name   = "terraform-test-1"
    region = "lon1"
    size   = "s-1vcpu-1gb"
    ssh_keys = [digitalocean_ssh_key.terraform.fingerprint]
}

resource "tls_private_key" "ssh_key" {
    algorithm = "ECDSA"
    ecdsa_curve = "P521"
}

resource "digitalocean_ssh_key" "terraform" {
    name       = "Terraform Test"
    public_key = tls_private_key.ssh_key.public_key_openssh
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

resource "linuxbox_ssh_authorized_key" "docker" {
    host_address = digitalocean_droplet.test.ipv4_address
    ssh_key = tls_private_key.ssh_key.private_key_pem
    key_to_add = file("~/.ssh/id_rsa.pub")
}

resource "linuxbox_docker_copy_image" "service" {
    depends_on = [linuxbox_docker.docker]
    host_address = digitalocean_droplet.test.ipv4_address
    ssh_key = tls_private_key.ssh_key.private_key_pem
    image_id = linuxbox_docker_build.sample_service.image_id
}

data "linuxbox_source_hash" "sample_service" {
    sources = ["sample_service"]
}

resource "linuxbox_docker_build" "sample_service" {
    source_dir = "${path.module}/sample_service"
    source_hash = data.linuxbox_source_hash.sample_service.hash
}

resource "linuxbox_docker_container" "webpage" {
    ssh_key = tls_private_key.ssh_key.private_key_pem
    image_id = linuxbox_docker_copy_image.service.image_id
    host_address = digitalocean_droplet.test.ipv4_address
    labels = {
        "traefik.enable" = "true"
        "traefik.port" = "80"
        "traefik.frontend.rule" = "Host:${digitalocean_droplet.test.ipv4_address}.xip.io"
    }
    
    env = {
        "foo" = "bar"
    }

    name = "nginx"

    restart = "unless-stopped"
}

resource "linuxbox_docker_container" "traefik" {
    
    depends_on = [linuxbox_docker.docker]

    ssh_key = tls_private_key.ssh_key.private_key_pem
    image_id = "traefik:v1.7.19-alpine"
    host_address = digitalocean_droplet.test.ipv4_address
    ports = [
        "80:80",
        "443:443",
    ]
    volumes = [
        "/var/run/docker.sock:/var/run/docker.sock",
        "/acme:/acme",
    ]

    restart = "unless-stopped"

    args = [
        "--accesslog",
        "--defaultentrypoints=http",
        "--defaultentrypoints=http,https",
        "--entrypoints=Name:http Address::80",        
        "--entryPoints=Name:https Address::443 TLS",
        "--docker",
        "--docker.watch",
        "--docker.exposedbydefault","false",
        "--acme",
        "--acme.entryPoint=https",
        "--acme.httpChallenge.entryPoint=http",
        "--acme.OnHostRule=true",
        "--acme.onDemand=false",
        "--acme.email=dragan@netice9.com",
        "--acme.tlsconfig=true",
        "--acme.storage=/acme/certs.json",
        // "--providers.docker.endpoint=unix:///var/run/docker.sock",
    ]

    name = "traefik"
}
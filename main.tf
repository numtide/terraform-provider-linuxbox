provider "digitalocean" {
}

provider "linuxbox" {
  ssh_session_limit = 3
}

resource "digitalocean_droplet" "test" {
  image              = "ubuntu-18-04-x64"
  name               = "terraform-test-1"
  region             = "lon1"
  size               = "s-1vcpu-1gb"
  private_networking = true
  ssh_keys           = [digitalocean_ssh_key.terraform.fingerprint]
}

resource "tls_private_key" "ssh_key" {
  algorithm   = "ECDSA"
  ecdsa_curve = "P521"
}

resource "digitalocean_ssh_key" "terraform" {
  name       = "Terraform Test"
  public_key = tls_private_key.ssh_key.public_key_openssh
}

resource "linuxbox_swap" "host_swap" {
  host_address = digitalocean_droplet.test.ipv4_address
  ssh_key      = tls_private_key.ssh_key.private_key_pem
  swap_size    = "100m"
}

# resource "linuxbox_docker" "docker" {
#     host_address = digitalocean_droplet.test.ipv4_address
#     ssh_key = tls_private_key.ssh_key.private_key_pem
# }

resource "linuxbox_directory" "foo" {
  host_address = digitalocean_droplet.test.ipv4_address
  ssh_key      = tls_private_key.ssh_key.private_key_pem
  path         = "/foo"
  owner        = 0
  group        = 0
}


resource "linuxbox_text_file" "test_file" {
  host_address = digitalocean_droplet.test.ipv4_address
  ssh_key      = tls_private_key.ssh_key.private_key_pem
  path         = "${linuxbox_directory.foo.path}/this is a test"
  content      = "testesttest"
  owner        = 0
  group        = 0
}

resource "linuxbox_binary_file" "test_binfile" {
  host_address   = digitalocean_droplet.test.ipv4_address
  ssh_key        = tls_private_key.ssh_key.private_key_pem
  path           = "${linuxbox_directory.foo.path}/this is a test - binary%"
  content_base64 = base64encode("test-binary")
  owner          = 0
  group          = 0
}

resource "linuxbox_run_setup" "install_docker" {
  host_address = digitalocean_droplet.test.ipv4_address
  ssh_key      = tls_private_key.ssh_key.private_key_pem

  setup = [
    "apt update",
    "apt install -y apt-transport-https ca-certificates curl software-properties-common",
    "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -",
    "add-apt-repository \"deb [arch=amd64] https://download.docker.com/linux/ubuntu bionic stable\"",
    "apt update",
    "apt install -y docker-ce",
  ]

}


resource "linuxbox_docker_network" "test_network" {
  depends_on   = [linuxbox_run_setup.install_docker]
  host_address = digitalocean_droplet.test.ipv4_address
  ssh_key      = tls_private_key.ssh_key.private_key_pem
  name         = "test"
}

resource "linuxbox_docker_run" "test_run" {
  depends_on        = [linuxbox_run_setup.install_docker]
  host_address      = digitalocean_droplet.test.ipv4_address
  ssh_key           = tls_private_key.ssh_key.private_key_pem
  image_id          = "alpine:latest"
  clear_entry_point = true
  args              = ["echo", "foo"]
}

resource "linuxbox_ssh_authorized_key" "docker" {
  host_address = digitalocean_droplet.test.ipv4_address
  ssh_key      = tls_private_key.ssh_key.private_key_pem
  key_to_add   = file("~/.ssh/id_rsa.pub")
}

resource "linuxbox_docker_copy_image" "service" {
  depends_on   = [linuxbox_run_setup.install_docker]
  host_address = digitalocean_droplet.test.ipv4_address
  ssh_key      = tls_private_key.ssh_key.private_key_pem
  image_id     = linuxbox_docker_build.sample_service.image_id
}

data "linuxbox_source_hash" "sample_service" {
  sources = ["sample_service"]
}

resource "linuxbox_docker_build" "sample_service" {
  source_dir  = "${path.module}/sample_service"
  source_hash = data.linuxbox_source_hash.sample_service.hash
}

resource "linuxbox_docker_container" "webpage" {
  depends_on = [linuxbox_run_setup.install_docker]

  ssh_key      = tls_private_key.ssh_key.private_key_pem
  image_id     = linuxbox_docker_copy_image.service.image_id
  host_address = digitalocean_droplet.test.ipv4_address
  labels = {
    "traefik.enable"        = "true"
    "traefik.port"          = "80"
    "traefik.frontend.rule" = "Host:${digitalocean_droplet.test.ipv4_address}.nip.io"
  }

  env = {
    "foo" = "bar"
  }

  name = "nginx"

  caps = ["IPC_LOCK"]

  restart = "unless-stopped"

  network = "bridge"

  log_opts = {
    "max-file" = 3
  }

  // 100 megs
  memory = 104857600
}

resource "linuxbox_docker_container" "traefik" {

  depends_on = [linuxbox_run_setup.install_docker]

  ssh_key      = tls_private_key.ssh_key.private_key_pem
  image_id     = "traefik:v1.7.19-alpine"
  host_address = digitalocean_droplet.test.ipv4_address
  ports = [
    "127.0.0.1:80:80",
    "443:443",
  ]
  volumes = [
    "/var/run/docker.sock:/var/run/docker.sock",
    "/acme:/acme",
  ]

  restart = "unless-stopped"

  privileged = true

  args = [
    "--accesslog",
    "--defaultentrypoints=http",
    "--defaultentrypoints=http,https",
    "--entrypoints=Name:http Address::80",
    "--entryPoints=Name:https Address::443 TLS",
    "--docker",
    "--docker.watch",
    "--docker.exposedbydefault", "false",
    "--acme",
    "--acme.entryPoint=https",
    "--acme.httpChallenge.entryPoint=http",
    "--acme.OnHostRule=true",
    "--acme.onDemand=false",
    "--acme.email=dragan@netice9.com",
    "--acme.tlsconfig=true",
    "--acme.storage=/acme/certs.json",
    # "--providers.docker.endpoint=unix:///var/run/docker.sock",
  ]

  name = "traefik"
}

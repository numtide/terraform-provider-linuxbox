provider "digitalocean" {
}


provider "linuxbox" {
}

# resource "digitalocean_droplet" "test" {
#     image  = "ubuntu-18-04-x64"
#     name   = "terraform-test-1"
#     region = "lon1"
#     size   = "s-1vcpu-1gb"
#     ssh_keys = [digitalocean_ssh_key.terraform.id]
# }

# resource "tls_private_key" "ssh_key" {
#     algorithm = "ECDSA"
#     # rsa_bits  = 2048
#     ecdsa_curve = "P521"
# }

# resource "digitalocean_ssh_key" "terraform" {
#     name       = "Terraform Test"
#     public_key = tls_private_key.ssh_key.public_key_openssh
# }

# resource "linuxbox_swap" "host_swap" {
#     host_address = digitalocean_droplet.test.ipv4_address
#     ssh_key = tls_private_key.ssh_key.private_key_pem
#     swap_size = "100m"
# }

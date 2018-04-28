resource "digitalocean_droplet" "info344" {
  image              = "docker-16-04"
  name               = "info344"
  region             = "SFO2"
  private_networking = false
  size               = "s-1vcpu-1gb"

  ssh_keys = ["${var.ssh_fingerprint}"]

  connection {
    user        = "root"
    type        = "ssh"
    private_key = "${file(var.pvt_key)}"
    timeout     = "2m"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo ufw 80",
      "sudo ufw 443",
      "sudo apt-get -y install letsencrypt",
      "sudo letsencrypt certonly --standalone -n --agree-tos --email ask710@uw.edu -d ask710.me",
    ]
  }
}

resource "digitalocean_domain" "ask710" {
  name       = "ask710.me"
  ip_address = "${digitalocean_droplet.info344.ipv4_address}"
}

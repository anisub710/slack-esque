resource "digitalocean_droplet" "info344-api" {
  image  = "docker"
  name   = "info344-api"
  region = "SFO2"
  size   = "1gb"

  ssh_keys = ["${var.ssh_fingerprint}"]

  connection {
    user        = "root"
    type        = "ssh"
    private_key = "${file(var.pvt_key)}"
    timeout     = "2m"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo apt-get install letsencrypt",
      "sudo letsencrypt certonly --standalone -d api.ask710.me",
    ]
  }
}

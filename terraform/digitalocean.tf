resource "digitalocean_project" "mybooks" {
  name        = "mybooks"
  description = "MyBooksRating"
  purpose     = "Web Application"
  environment = "Production"
  resources = [digitalocean_droplet.www.urn]
}

data "digitalocean_volume" "block-volume" {
  name   = var.digitalocean_volume_name
  region = var.digitalocean_droplet_region
}

resource "digitalocean_droplet" "www" {
  image = "ubuntu-20-04-x64"
  name = "mybooks-droplet"
  region = var.digitalocean_droplet_region
  size = "s-1vcpu-2gb"
  private_networking = true
  ssh_keys = [
    data.digitalocean_ssh_key.terraform.id
  ]
  user_data = templatefile("${path.root}/www-cloud-init.yaml", {
    "PWD" = "$${PWD}",
    "certbot_email" = var.certbot_email,
    "mysql_user" = var.mysql_user,
    "mysql_password" = var.mysql_password,
    "mybooks_dns" = var.mybooks_dns,
    "cloudflare_email" = var.cloudflare_email,
    "cloudflare_api_key" = var.cloudflare_api_key,
    "cloudflare_domain" = var.cloudflare_domain,
    "digitalocean_volume_name" = var.digitalocean_volume_name,
    "google_client_id" = var.google_client_id,
    "google_client_secret" = var.google_client_secret,
    "google_api_key" = var.google_api_key
  })
  connection {
    host = self.ipv4_address
    user = "root"
    type = "ssh"
    private_key = file(var.pvt_key)
    timeout = "2m"
  }
//  provisioner "remote-exec" {
//    inline = [
//      "export PATH=$PATH:/usr/bin",
//      # install nginx
//      "sudo apt-get update",
//      "sudo apt-get -y install nginx",
//      # install mysql
//      "sudo rm /var/lib/mysql/ -R",
//      "sudo rm /etc/mysql/ -R",
//      "sudo apt-get autoremove mysql* --purge",
//      "wget https://dev.mysql.com/get/Downloads/MySQL-5.5/mysql-5.5.56-linux-glibc2.5-x86_64.tar.gz",
//      "sudo groupadd mysql",
//      "sudo useradd -g mysql mysql",
//      "sudo tar -xvf mysql-5.5.56-linux-glibc2.5-x86_64.tar.gz",
//      "sudo mv mysql-5.5.56-linux-glibc2.5-x86_64 /usr/local/",
//      "cd /usr/local",
//      "sudo mv mysql-5.5.56-linux-glibc2.5-x86_64 mysql",
//      "cd mysql",
//      "sudo chown -R mysql:mysql *",
//      "sudo apt-get install libaio1",
//      "sudo scripts/mysql_install_db --user=mysql",
//      "sudo chown -R root",
//      "sudo chown -R mysql data",
//      "sudo cp support-files/my-medium.cnf /etc/my.cnf",
//      "sudo bin/mysqld_safe --user=mysql & sudo cp support-files/mysql.server /etc/init.d/mysql.server",
//      "sudo bin/mysqladmin -u root password 'mybooks2021!'",
//      "sudo ln -s /usr/local/mysql/bin/mysql /usr/local/bin/mysql",
//      "sudo /etc/init.d/mysql.server start",
//      "sudo /etc/init.d/mysql.server status",
//      "sudo update-rc.d -f mysql.server defaults"
//
//    ]
//  }

}

resource "digitalocean_volume_attachment" "vol-attachment" {
  droplet_id = digitalocean_droplet.www.id
  volume_id  = data.digitalocean_volume.block-volume.id
}

resource "digitalocean_firewall" "www" {
  name = "only-22-80-and-443"
  droplet_ids = [digitalocean_droplet.www.id]

  inbound_rule {
    protocol    = "tcp"
    port_range  = "443"
    source_addresses = data.cloudflare_ip_ranges.cloudflare.cidr_blocks
  }
  inbound_rule {
    protocol    = "tcp"
    port_range  = "80"
    source_addresses = data.cloudflare_ip_ranges.cloudflare.cidr_blocks
  }
  inbound_rule {
    protocol    = "tcp"
    port_range  = "22"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }
  inbound_rule {
    protocol    = "icmp"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol    = "tcp"
    port_range = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol    = "udp"
    port_range = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol    = "icmp"
    port_range = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}

data "cloudflare_ip_ranges" "cloudflare" {}


data "digitalocean_droplet" "www" {
  name = "mybooks-droplet"
  depends_on = [digitalocean_droplet.www]
}
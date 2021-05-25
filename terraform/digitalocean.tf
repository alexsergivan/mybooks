
data "digitalocean_volume" "block-volume" {
  name   = var.digitalocean_volume_name
  region = var.digitalocean_droplet_region
}

resource "digitalocean_droplet" "www" {
  image = "ubuntu-20-04-x64"
  name = "mybooks-www"
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
    "google_api_key" = var.google_api_key,
    "spaces_key" = var.access_id,
    "spaces_secret" = var.secret_key
  })
  connection {
    host = self.ipv4_address
    user = "root"
    type = "ssh"
    private_key = file(var.pvt_key)
    timeout = "2m"
  }
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
  name = "mybooks-www"
  depends_on = [digitalocean_droplet.www]
}


resource "digitalocean_spaces_bucket" "mybooks-static-bucket" {
  name   = "mybooks-static-bucket"
  region = "fra1"
  acl    = "public-read"
}

# Add a CDN endpoint to the Spaces Bucket
resource "digitalocean_cdn" "mycdn" {
  origin = digitalocean_spaces_bucket.mybooks-static-bucket.bucket_domain_name
}

# Output the endpoint for the CDN resource
output "fqdn" {
  value = digitalocean_cdn.mycdn.endpoint
}

resource "digitalocean_project" "mybooks-prod" {
  name        = "mybooks-prod"
  description = "MyBooksRating"
  purpose     = "Web Application"
  environment = "Production"
  resources = [
    digitalocean_droplet.www.urn,
    digitalocean_spaces_bucket.mybooks-static-bucket.urn
  ]
}
terraform {
  required_providers {
    digitalocean = {
      source = "digitalocean/digitalocean"
      version = "1.22.2"
    }
    cloudflare = {
      source = "cloudflare/cloudflare"
      version = "~> 2.0"
    }
  }
}

provider "digitalocean" {
  token = var.do_token
}

provider "cloudflare" {
  api_key = var.cloudflare_api_key
  email = var.cloudflare_email
}

data "digitalocean_ssh_key" "terraform" {
  name = "oleks"
}
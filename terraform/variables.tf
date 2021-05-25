variable "cloudflare_domain" {}
variable "mybooks_dns" {}

variable "do_token" {}
variable "pvt_key" {}

variable "digitalocean_volume_name" {}
variable "digitalocean_droplet_region" {}
variable "certbot_email" {}
variable "mysql_user" {}
variable "mysql_password" {}

variable "cloudflare_email" {
  type = string
}

variable "cloudflare_api_key" {
  type = string
}

variable "cloudflare_api_token" {
  type = string
}

variable "google_client_id" {}
variable "google_client_secret" {}
variable "google_api_key" {}

# DO Spaces.
variable "access_id" {}
variable "secret_key" {}
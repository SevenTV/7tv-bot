data "terraform_remote_state" "infra" {
  backend = "remote"

  config = {
    organization = "7tv"
    workspaces   = {
      name = local.infra_workspace_name
    }
  }
}

variable "region" {
  type    = string
  default = "us-east-2"
}

variable "namespace" {
  type    = string
  default = "app"
}

variable "image_pull_policy" {
  type    = string
  default = "Always"
}

variable "oauth_secret" {
  type = string
  default = "twitch-irc-oauth"
}

variable "twitch_client_id" {
  type    = string
  default = ""
}

variable "twitch_client_secret" {
  type    = string
  default = ""
}

variable "twitch_username" {
  type    = string
  default = ""
}

variable "ratelimit_join" {
  type    = string
  default = ""
}

variable "ratelimit_auth" {
  type    = string
  default = ""
}

variable "ratelimit_reset" {
  type    = string
  default = ""
}

variable "nats_irc_raw_subject" {
  type    = string
  default = ""
}

variable "nats_bot_api_subject" {
  type    = string
  default = ""
}

variable "nats_twitch_irc_stream" {
  type    = string
  default = ""
}

variable "mongo_bot_database" {
  type    = string
  default = ""
}

variable "mongo_bot_users_collection" {
  type    = string
  default = ""
}

variable "mongo_bot_global_stats_collection" {
  type    = string
  default = ""
}
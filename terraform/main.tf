terraform {

  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "7tv"

    workspaces {
      prefix = "seventv-stats-"
    }
  }
}

locals {
  infra_workspace_name = replace(terraform.workspace, "stats", "infra")
  infra                = data.terraform_remote_state.infra.outputs
  image_url_template   = format("ghcr.io/seventv/7tv-bot/#APP:%s-latest", trimprefix(terraform.workspace, "seventv-stats-"))
}

module "oauth" {
  source = "./oauth"

  oauth_secret         = var.oauth_secret
  twitch_client_id     = var.twitch_client_id
  twitch_client_secret = var.twitch_client_secret
  image_url_template   = local.image_url_template
}

module "irc-reader" {
  source                     = "./reader"
  oauth_secret               = var.oauth_secret
  twitch_username            = var.twitch_username
  ratelimit_join             = var.ratelimit_join
  ratelimit_auth             = var.ratelimit_auth
  ratelimit_reset            = var.ratelimit_reset
  nats_irc_raw_subject       = var.nats_irc_raw_subject
  nats_bot_api_subject       = var.nats_bot_api_subject
  nats_twitch_irc_stream     = var.nats_twitch_irc_stream
  mongo_bot_database         = var.mongo_bot_database
  mongo_bot_users_collection = var.mongo_bot_users_collection
  infra                      = local.infra
}
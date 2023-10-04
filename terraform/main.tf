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
  image_url_template     = format("ghcr.io/seventv/7tv-bot/#APP:%s-latest", trimprefix(terraform.workspace, "seventv-stats-"))
}

module "oauth" {
  source = "./oauth"

  twitch_client_id     = var.twitch_client_id
  twitch_client_secret = var.twitch_client_secret
  image_url_template     = local.image_url_template
}

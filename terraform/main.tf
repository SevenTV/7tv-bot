terraform {
  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "seventv-stats"
  }
}

locals {
  app                  = variable.app
  infra_workspace_name = replace(terraform.workspace, "stats", "infra")
  infra                = data.terraform_remote_state.infra.outputs
  image_url_prefix     = format("ghcr.io/seventv/7tv-bot/%s-%s-latest", variable.app, trimprefix(terraform.workspace, "seventv-stats"))
}

module "oauth" {
  source = join("", "./", local.app)

  twitch_client_id     = var.twitch_client_id
  twitch_client_secret = var.twitch_client_secret
  image_url_prefix     = local.image_url_prefix
}

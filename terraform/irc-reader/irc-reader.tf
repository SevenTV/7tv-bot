terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.7.0"
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.18.1"
    }
  }
}

data "kubernetes_namespace" "app" {
  metadata {
    name = var.namespace
  }
}

data "kubernetes_secret" "oauth" {
  metadata {
    name = var.oauth_secret
    namespace = var.namespace
  }
  binary_data = {
    "access-token" = ""
  }
}

// Define config secret
resource "kubernetes_secret" "app" {
  metadata {
    name      = "stats-irc-reader"
    namespace = var.namespace
  }
  depends_on = [data.kubernetes_secret.oauth]

  data = {
    "config.yaml" = templatefile("${path.module}/config.template.yaml", {
      twitch_username  = var.twitch_username
      twitch_oauth     = data.kubernetes_secret.oauth.binary_data["access-token"]
      ratelimit_join   = var.ratelimit_join
      ratelimit_auth   = var.ratelimit_auth
      ratelimit_reset  = var.ratelimit_reset
      redis_username   = "default"
      redis_password   = var.infra.redis_password
      redis_address    = var.infra.redis_host
      redis_sentinel   = true
      redis_master     = "mymaster"
      redis_database   = 5
      nats_url         = "nats.database.svc.cluster.local:4222"
      nats_stream      = var.nats_twitch_irc_stream
      nats_irc_raw     = var.nats_irc_raw_subject
      nats_bot_api     = var.nats_bot_api_subject
      mongo_uri        = var.infra.mongodb_uri
      mongo_username   = var.infra.mongodb_user_app.username
      mongo_password   = var.infra.mongodb_user_app.password
      mongo_database   = var.mongo_bot_database
      mongo_collection = var.mongo_bot_users_collection
    })
  }
}
resource "kubernetes_service" "app" {
  metadata {
    name   = "stats-irc-reader"
    labels = {
      app = "stats-irc-reader"
    }
  }
  spec {
    selector = {
      app = "stats-irc-reader"
    }
    cluster_ip = "None"
  }
}

resource "kubernetes_stateful_set" "app" {
  metadata {
    name      = "stats-irc-reader"
    namespace = data.kubernetes_namespace.app.metadata[0].name
    labels    = {
      app = "stats-irc-reader"
    }
  }

  lifecycle {
    replace_triggered_by = [kubernetes_secret.app]
  }

  timeouts {
    create = "5m"
    update = "5m"
    delete = "5m"
  }

  spec {
    selector {
      match_labels = {
        app = "stats-irc-reader"
      }
    }

    replicas     = 1
    service_name = kubernetes_service.app.metadata[0].name

    template {
      metadata {
        labels = {
          app = "stats-irc-reader"
        }
      }

      spec {
        container {
          name  = "stats-irc-reader"
          image = replace(var.image_url_template, "#APP", "irc-reader")

          resources {
            limits = {
              cpu    = "1"
              memory = "1Gi"
            }

            requests = {
              cpu    = "10m"
              memory = "50Mi"
            }
          }

          volume_mount {
            name       = "config"
            mount_path = "/app/config.yaml"
            sub_path   = "config.yaml"
          }
        }

        volume {
          name = "config"

          secret {
            secret_name = kubernetes_secret.app.metadata[0].name
          }
        }
      }
    }
  }
}

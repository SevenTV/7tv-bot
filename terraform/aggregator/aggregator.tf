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

// Define config secret
resource "kubernetes_secret" "app" {
  metadata {
    name      = "stats-aggregator"
    namespace = var.namespace
  }

  data = {
    "config.yaml" = templatefile("${path.module}/config.template.yaml", {
      max_workers      = 6
      nats_consumer    = "irc-stats-aggregator"
      nats_url         = "nats.database.svc.cluster.local:4222"
      nats_irc_raw     = var.nats_irc_raw_subject
      nats_stream      = var.nats_twitch_irc_stream
      mongo_uri        = var.infra.mongodb_uri
      mongo_username   = var.infra.mongodb_user_app.username
      mongo_password   = var.infra.mongodb_user_app.password
      mongo_database   = var.mongo_bot_database
      mongo_collection = var.mongo_bot_global_stats_collection
    })
  }
}

resource "kubernetes_deployment" "app" {
  metadata {
    name      = "stats-aggregator"
    namespace = data.kubernetes_namespace.app.metadata[0].name
    labels    = {
      k8slens-edit-resource-version = "v1"
      app = "stats-aggregator"
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
        app = "stats-aggregator"
      }
    }

    replicas = 3

    template {
      metadata {
        labels = {
          app = "stats-aggregator"
        }
      }

      spec {
        container {
          name  = "stats-aggregator"
          image = replace(var.image_url_template, "#APP", "aggregator")
          image_pull_policy = "Always"

          resources {
            limits = {
              cpu    = "2"
              memory = "2Gi"
            }

            requests = {
              cpu    = "10m"
              memory = "2Gi"
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

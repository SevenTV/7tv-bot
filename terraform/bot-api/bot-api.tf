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
    name      = "stats-bot-api"
    namespace = var.namespace
  }

  data = {
    "config.yaml" = templatefile("${path.module}/config.template.yaml", {
      nats_url         = "nats.database.svc.cluster.local:4222"
      nats_bot_api     = var.nats_bot_api_subject
      port             = "7777"
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
    name   = "stats-bot-api"
    labels = {
      app = "stats-bot-api"
    }
  }
  spec {
    selector = {
      app = "stats-bot-api"
    }

    port {
      name       = "http"
      port       = 7777
      target_port = 7777
      protocol   = "TCP"
    }
  }
}

resource "kubernetes_deployment" "app" {
  metadata {
    name      = "stats-bot-api"
    namespace = data.kubernetes_namespace.app.metadata[0].name
    labels    = {
      app = "stats-bot-api"
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
        app = "stats-bot-api"
      }
    }

    replicas = 1

    template {
      metadata {
        labels = {
          app = "stats-bot-api"
        }
      }

      spec {
        container {
          name  = "stats-bot-api"
          image = replace(var.image_url_template, "#APP", "bot-api")

          port {
            name           = "http"
            container_port = 7777
            protocol       = "TCP"
          }

          resources {
            limits = {
              cpu    = "500m"
              memory = "512Mi"
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

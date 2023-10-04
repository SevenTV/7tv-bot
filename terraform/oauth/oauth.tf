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
    name      = "stats-oauth"
    namespace = var.namespace
  }

  data = {
    "config.yaml" = templatefile("${path.module}/config.template.yaml", {
      redirect_uri  = "http://localhost:7777"
      namespace     = var.namespace
      oauth_secret  = "twitch-irc-oauth"
      port          = 7777
      client_id     = var.twitch_client_id
      client_secret = var.twitch_client_secret
    })
  }
}

resource "kubernetes_deployment" "app" {
  metadata {
    name      = "stats-oauth"
    namespace = data.kubernetes_namespace.app.metadata[0].name
    labels = {
      app = "stats-oauth"
    }
  }

  lifecycle {
    replace_triggered_by = [kubernetes_secret.app]
  }

  timeouts {
    create = "10m"
    update = "10m"
    delete = "10m"
  }

  spec {
    selector {
      match_labels = {
        app = "stats-oauth"
      }
    }

    replicas = 1

    template {
      metadata {
        labels = {
          app = "stats-oauth"
        }
      }

      spec {
        container {
          name  = "stats-oauth"
          image = replace(var.image_url_template, "#APP", "oauth")

          port {
            name           = "http"
            container_port = 7777
            protocol       = "TCP"
          }

          resources {
            limits = {
              cpu    = "50m"
              memory = "128Mi"
            }

            requests = {
              cpu    = "10m"
              memory = "20Mi"
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

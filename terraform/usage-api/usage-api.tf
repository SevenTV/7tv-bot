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

resource "kubernetes_secret" "app" {
  metadata {
    name      = "stats-usage-api"
    namespace = var.namespace
  }

  data = {
    "config.yaml" = templatefile("${path.module}/config.template.yaml", {
      nats_url          = "nats.database.svc.cluster.local:4222"
      nats_topic_emotes = var.nats_emotes_global
      port              = "7777"
      mongo_uri         = var.infra.mongodb_uri
      mongo_username    = var.infra.mongodb_user_app.username
      mongo_password    = var.infra.mongodb_user_app.password
      mongo_database    = var.mongo_bot_database
      mongo_collection  = var.mongo_bot_global_stats_collection
    })
  }
}

resource "kubernetes_service" "app" {
  metadata {
    name   = "stats-usage-api"
    labels = {
      app = "stats-usage-api"
    }
  }
  spec {
    selector = {
      app = "stats-usage-api"
    }
    port {
      name        = "http"
      port        = 7777
      target_port = "rest"
      protocol    = "TCP"
    }
  }
}

resource "kubernetes_ingress_v1" "app" {
  metadata {
    name      = "emote-usage-api"
    namespace = data.kubernetes_namespace.app.metadata[0].name
    annotations = {
      "kubernetes.io/ingress.class"                         = "nginx"
      "external-dns.alpha.kubernetes.io/target"             = var.infra.cloudflare_tunnel_hostname.regular
      "external-dns.alpha.kubernetes.io/cloudflare-proxied" = "true"
      "nginx.ingress.kubernetes.io/limit-connections" : "64"
      "nginx.ingress.kubernetes.io/proxy-body-size" : "7m"
    }
  }

  spec {
    rule {
      // set "stats" as subdomain
      host = join(".", ["stats", var.infra.secondary_zone])
      http {
        // Developer Portal
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = kubernetes_service.app.metadata[0].name
              port {
                name = "rest"
              }
            }
          }
        }
      }
    }
  }
}

resource "kubernetes_deployment" "app" {
  metadata {
    name      = "stats-usage-api"
    namespace = data.kubernetes_namespace.app.metadata[0].name
    labels    = {
      app = "stats-usage-api"
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
        app = "stats-usage-api"
      }
    }

    replicas = 1

    template {
      metadata {
        labels = {
          app = "stats-usage-api"
        }
      }
      spec {
        container {
          name              = "stats-usage-api"
          image             = replace(var.image_url_template, "#APP", "bot-api")
          image_pull_policy = "Always"

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
              memory = "64Mi"
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
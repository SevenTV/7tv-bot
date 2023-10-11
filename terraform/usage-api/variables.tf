

variable "namespace" {
  type    = string
  default = "app"
}

variable "nats_emotes_global" {
  type    = string
  default = ""
}

variable "mongo_bot_database" {
  type    = string
  default = ""
}

variable "mongo_bot_global_stats_collection" {
  type    = string
  default = ""
}

variable "image_url_template" {
  type    = string
  default = ""
}

variable "infra" {
  type = any
}
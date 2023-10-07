

variable "namespace" {
  type    = string
  default = "app"
}

variable "nats_bot_api_subject" {
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

variable "image_url_template" {
  type    = string
  default = ""
}

variable "infra" {
  type = any
}
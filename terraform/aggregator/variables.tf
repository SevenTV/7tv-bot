

variable "namespace" {
  type    = string
  default = "app"
}

variable "nats_irc_raw_subject" {
  type    = string
  default = ""
}

variable "nats_twitch_irc_stream" {
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
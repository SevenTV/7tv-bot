

variable "namespace" {
  type    = string
  default = "app"
}

variable "oauth_secret" {
  type = string
  default = ""
}

variable "twitch_username" {
  type    = string
  default = ""
}

variable "twitch_oauth" {
  type    = string
  default = ""
}

variable "ratelimit_join" {
  type    = string
  default = ""
}

variable "ratelimit_auth" {
  type    = string
  default = ""
}

variable "ratelimit_reset" {
  type    = string
  default = ""
}

variable "nats_irc_raw_subject" {
  type    = string
  default = ""
}

variable "nats_bot_api_subject" {
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
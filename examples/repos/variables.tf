variable "oracle_hostname" {
  description = "The hostname for the Oracle database."
  type        = string
  default     = "localhost"
}

variable "oracle_port" {
  description = "The port for the Oracle database."
  type        = number
  default     = 1521
}

variable "iam_role" {
  description = "The IAM role for AWS Secrets Manager."
  type        = string
  default = ""
}

variable "secrets_path" {
  description = "The path in AWS Secrets Manager where secrets are stored."
  type        = string
}

variable "api_key" {
  description = "The API key for Altr."
  type        = string
  sensitive = false
}

variable "secret" {
  description = "The secret for Altr."
  type        = string
  sensitive = true
}

variable "org_id" {
  description = "The organization ID for Altr."
  type        = string
  sensitive = false
}

variable "base_url" {
  description = "The base URL for Altr."
  type        = string
  sensitive = false
}

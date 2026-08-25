variable "aws_region" {
  description = "AWS region to deploy into"
  type        = string
  default     = "us-east-1"
}

variable "instance_type" {
  description = "EC2 instance type. t3.small (2GB RAM) - modernc.org/sqlite (the pure-Go SQLite driver) is a large transpiled package and compiling it inside Docker exhausted t3.micro's 1GB RAM on the bank-demo project, so this defaults straight to t3.small this time."
  type        = string
  default     = "t3.small"
}

variable "github_repo_url" {
  description = "Public HTTPS URL of the sentinel GitHub repository, e.g. https://github.com/your-username/sentinel.git"
  type        = string
}

variable "ssh_allowed_cidr" {
  description = "CIDR block allowed to SSH into the instance. Defaults to open (0.0.0.0/0) for simplicity in a demo account - restrict to your own IP/32 for tighter security."
  type        = string
  default     = "0.0.0.0/0"
}

variable "project_name" {
  description = "Name prefix used for tagging AWS resources"
  type        = string
  default     = "sentinel"
}

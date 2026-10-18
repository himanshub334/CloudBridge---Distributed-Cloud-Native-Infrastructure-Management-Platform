variable "aws_region" { type = string; default = "us-east-1" }
variable "project_name" { type = string; default = "cloudbridge-demo" }
variable "state_bucket_name" { type = string; description = "Globally unique bucket name for Terraform state" }

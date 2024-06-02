terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-west-1"

  default_tags {
    tags = {
      deployment = "OpenTofu"
      project    = "picloud"
      repo       = "https://github.com/fullerzz/picloud"
    }
  }
}

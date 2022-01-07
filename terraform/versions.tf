terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.51"
    }
  }

  backend "s3" {
    key    = "gateway-lb"
    region = "eu-central-1"
  }

  required_version = ">= 1.0.10"
}

provider "aws" {
  region = "eu-central-1"
}

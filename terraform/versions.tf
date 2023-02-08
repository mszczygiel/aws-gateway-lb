terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
  backend "local" {
    path = "terraform.tfstate"
  }

  required_version = ">= 1.3.7"
}

provider "aws" {
  region = "eu-central-1"
}

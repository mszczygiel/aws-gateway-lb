resource "aws_key_pair" "default" {
  public_key = file("${path.module}/pubkey")
}

data "aws_caller_identity" "current" {}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.51"
    }
  }

  backend "s3" {
    bucket = "mszczygiel-playground-tfstate"
    key    = "gateway-lb"
    region = "eu-central-1"
  }

  required_version = ">= 1.0.10"
}

provider "aws" {
  region  = "eu-central-1"
}

data "aws_caller_identity" "current" {}

locals {
  az = "eu-central-1a"
  geneve_port = 6081
  init_intance = <<EOF
yum -y update && yum -y install nc
echo "TERM=vt100" | tee -a /etc/environment
  EOF
}

resource "aws_key_pair" "default" {
    public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCoZnn82NlaDWzzCzHT/86oofHKatrKTx3GnP3cnGgyO2KA3NH0naYlZsaUISfT3imoYNqtnKdRwNWfH3sPtCAtzLoCDoQwyL5aIjZIKdzjcEVPxHcW2B+sEIaKu30KmHTfZtT6aFL1/JXlrlGMy/c1IA0QteI+pxQOHQJf+b3d7FjrCz4SJlmh5Lseslh319r69RVQ6MuN435uJJrawywvGsuB6dzYoDMt0Y1lSUeREr3L1pHq9VjTvfMhF3FVMyPcx5zuDAQY1XgMiDvV2NVl2CTdNPq3Z9o/BDfxYDfytOL/5Rrs/QxVA9LuQAUx+6yJr2t6HBb95uWcls89aUV2Y0lTm32c7iGfvePrADP9j9tQJqvhHkfDUk7prR6w/HIUkCAUPjqvrITUP6c5sCUICNEvjSmbdo6NMtzWmt3zvp0Z2SMTIhjIpxNbvUoWfMPTEKMTgzQDCii8G5BFGQYsqxJwv6tZy+/5al48WFWOpdeF4hS7AA0HyQPJQXpKbO8="
}

resource "aws_vpc" "main" {
    cidr_block = "192.168.0.0/16"
}

resource "aws_internet_gateway" "gw" {
    vpc_id = aws_vpc.main.id
}

resource "aws_subnet" "apps_a" {
    vpc_id = aws_vpc.main.id
    cidr_block = "192.168.1.0/24"
    availability_zone = local.az
}
resource "aws_subnet" "apps_b" {
    vpc_id = aws_vpc.main.id
    cidr_block = "192.168.2.0/24"
    availability_zone = local.az
}
resource "aws_subnet" "endpoint" {
    vpc_id = aws_vpc.main.id
    cidr_block = "192.168.10.0/24"
    availability_zone = local.az
}
resource "aws_subnet" "appliances" {
    vpc_id = aws_vpc.main.id
    cidr_block = "192.168.20.0/24"
    availability_zone = local.az
}

resource "aws_route_table" "appliances" {
    vpc_id = aws_vpc.main.id

    route {
        cidr_block = aws_subnet.appliances.cidr_block
        vpc_endpoint_id = aws_vpc_endpoint.censored_communication.id
    }
    route {
        cidr_block = "0.0.0.0/0"
        gateway_id = aws_internet_gateway.gw.id
    }
}

resource "aws_route_table_association" "appliances" {
  subnet_id = aws_subnet.appliances.id
  route_table_id = aws_route_table.appliances.id
}

resource "aws_route_table" "apps_a" {
    vpc_id = aws_vpc.main.id

    route {
        cidr_block = aws_subnet.apps_b.cidr_block
        vpc_endpoint_id = aws_vpc_endpoint.censored_communication.id
    }
    route {
        cidr_block = "0.0.0.0/0"
        gateway_id = aws_internet_gateway.gw.id
    }
}

resource "aws_route_table_association" "apps_a" {
  subnet_id = aws_subnet.apps_a.id
  route_table_id = aws_route_table.apps_a.id
}

resource "aws_route_table" "apps_b" {
    vpc_id = aws_vpc.main.id

    route {
        cidr_block = aws_subnet.apps_a.cidr_block
        vpc_endpoint_id = aws_vpc_endpoint.censored_communication.id
    }

    route {
        cidr_block = "0.0.0.0/0"
        gateway_id = aws_internet_gateway.gw.id
    }
}

resource "aws_route_table_association" "apps_b" {
  subnet_id = aws_subnet.apps_b.id
  route_table_id = aws_route_table.apps_b.id
}

resource "aws_security_group" "permissive_egress" {
    vpc_id = aws_vpc.main.id
    egress {
        from_port = 0
        to_port = 0
        protocol = "-1"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_security_group" "allow_ssh" {
    vpc_id = aws_vpc.main.id
    ingress {
        from_port = 22
        to_port = 22
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_security_group" "allow_geneve" {
    vpc_id = aws_vpc.main.id
    ingress {
        from_port = local.geneve_port
        to_port = local.geneve_port
        protocol = "udp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_instance" "app_a" {
  ami           = "ami-00051469f31042765"
  instance_type = "t2.micro"
  key_name = aws_key_pair.default.key_name
  vpc_security_group_ids = [aws_security_group.allow_ssh.id, aws_security_group.permissive_egress.id]
  availability_zone = local.az
  associate_public_ip_address = true
  subnet_id = aws_subnet.apps_a.id
  private_ip = "192.168.1.10"
  user_data = local.init_intance
}
resource "aws_instance" "app_b" {
  ami           = "ami-00051469f31042765"
  instance_type = "t2.micro"
  key_name = aws_key_pair.default.key_name
  vpc_security_group_ids = [aws_security_group.allow_ssh.id, aws_security_group.permissive_egress.id]
  availability_zone = local.az
  associate_public_ip_address = true
  subnet_id = aws_subnet.apps_b.id
  private_ip = "192.168.2.10"
  user_data = local.init_intance
}
resource "aws_instance" "appliance" {
  ami           = "ami-00051469f31042765"
  instance_type = "t2.micro"
  key_name = aws_key_pair.default.key_name
  vpc_security_group_ids = [aws_security_group.allow_ssh.id, aws_security_group.allow_geneve.id, aws_security_group.permissive_egress.id]
  availability_zone = local.az
  associate_public_ip_address = true
  subnet_id = aws_subnet.appliances.id
  private_ip = "192.168.20.10"
  user_data = local.init_intance
}

resource "aws_lb" "gateway" {
    load_balancer_type = "gateway"
    subnets = [aws_subnet.appliances.id]
}

resource "aws_lb_target_group" "appliances" {
    port = local.geneve_port
    protocol = "GENEVE"
    vpc_id = aws_vpc.main.id
}

resource "aws_lb_target_group_attachment" "appliances" {
    target_group_arn = aws_lb_target_group.appliances.arn
    target_id = aws_instance.appliance.id
    port = local.geneve_port
}

resource "aws_lb_listener" "gateway" {
    load_balancer_arn = aws_lb.gateway.arn
    default_action {
        target_group_arn = aws_lb_target_group.appliances.id
        type = "forward"
    }
}

resource "aws_vpc_endpoint_service" "censor_service" {
    acceptance_required = false
    allowed_principals = [data.aws_caller_identity.current.arn]
    gateway_load_balancer_arns = [aws_lb.gateway.arn]
}

resource "aws_vpc_endpoint" "censored_communication" {
    service_name = aws_vpc_endpoint_service.censor_service.service_name
    subnet_ids = [aws_subnet.endpoint.id]
    vpc_endpoint_type = aws_vpc_endpoint_service.censor_service.service_type
    vpc_id = aws_vpc.main.id
}

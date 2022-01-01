resource "aws_vpc" "apps" {
  cidr_block = "192.168.0.0/16"
}

# resource "aws_s3_bucket" "apps_flow_logs" {
#     bucket = "mszczygiel-demo-apps-flow-logs"
#     acl = "log-delivery-write"
#     force_destroy = true
# }

# resource "aws_flow_log" "apps" {
#   log_destination      = aws_s3_bucket.apps_flow_logs.arn
#   log_destination_type = "s3"
#   traffic_type         = "ALL"
#   vpc_id               = aws_vpc.apps.id
# }

resource "aws_internet_gateway" "apps_igw" {
  vpc_id = aws_vpc.apps.id
}

resource "aws_subnet" "apps_a" {
  vpc_id            = aws_vpc.apps.id
  cidr_block        = "192.168.1.0/24"
  availability_zone = local.az
}

resource "aws_subnet" "endpoint" {
  vpc_id            = aws_vpc.apps.id
  cidr_block        = "192.168.10.0/24"
  availability_zone = local.az
}



resource "aws_route_table" "apps_a" {
  vpc_id = aws_vpc.apps.id

  route {
    cidr_block      = "0.0.0.0/0"
    vpc_endpoint_id = aws_vpc_endpoint.censored_communication.id
  }
}

resource "aws_route_table_association" "apps_a" {
  subnet_id      = aws_subnet.apps_a.id
  route_table_id = aws_route_table.apps_a.id
}

resource "aws_route_table" "igw" {
  vpc_id = aws_vpc.apps.id

  route {
    cidr_block      = aws_subnet.apps_a.cidr_block
    vpc_endpoint_id = aws_vpc_endpoint.censored_communication.id
  }
}

resource "aws_route_table_association" "igw" {
  gateway_id     = aws_internet_gateway.apps_igw.id
  route_table_id = aws_route_table.igw.id
}


resource "aws_route_table" "endpoint" {
  vpc_id = aws_vpc.apps.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.apps_igw.id
  }
}

resource "aws_route_table_association" "endpoint" {
  subnet_id      = aws_subnet.endpoint.id
  route_table_id = aws_route_table.endpoint.id
}

resource "aws_vpc_endpoint" "censored_communication" {
  service_name      = aws_vpc_endpoint_service.censor_service.service_name
  subnet_ids        = [aws_subnet.endpoint.id]
  vpc_endpoint_type = aws_vpc_endpoint_service.censor_service.service_type
  vpc_id            = aws_vpc.apps.id
}

resource "aws_security_group" "apps_allow_ssh" {
  vpc_id = aws_vpc.apps.id
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
resource "aws_security_group" "apps_allow_chat" {
  vpc_id = aws_vpc.apps.id

  ingress {
    from_port   = local.chat_port
    to_port     = local.chat_port
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "apps_permissive_egress" {
  vpc_id = aws_vpc.apps.id
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "app_a" {
  ami                         = "ami-00051469f31042765"
  instance_type               = "t2.micro"
  key_name                    = aws_key_pair.default.key_name
  vpc_security_group_ids      = [aws_security_group.apps_allow_ssh.id, aws_security_group.apps_permissive_egress.id, aws_security_group.apps_allow_chat.id]
  availability_zone           = local.az
  associate_public_ip_address = true
  subnet_id                   = aws_subnet.apps_a.id
  user_data                   = local.init_intance
}

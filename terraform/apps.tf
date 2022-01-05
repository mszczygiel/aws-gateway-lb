resource "aws_vpc_endpoint" "censored_communication" {
  service_name      = aws_vpc_endpoint_service.censor_service.service_name
  subnet_ids        = [aws_subnet.endpoint.id]
  vpc_endpoint_type = aws_vpc_endpoint_service.censor_service.service_type
  vpc_id            = aws_vpc.main.id
}

resource "aws_vpc_endpoint" "censored_communication_public" {
  service_name      = aws_vpc_endpoint_service.censor_service.service_name
  subnet_ids        = [aws_subnet.endpoint_public.id]
  vpc_endpoint_type = aws_vpc_endpoint_service.censor_service.service_type
  vpc_id            = aws_vpc.main.id
}

resource "aws_security_group" "apps_allow_ssh" {
  vpc_id = aws_vpc.main.id
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
resource "aws_security_group" "apps_allow_chat_local" {
  vpc_id = aws_vpc.main.id

  ingress {
    from_port   = local.chat_port
    to_port     = local.chat_port
    protocol    = "tcp"
    cidr_blocks = [aws_subnet.apps_a.cidr_block, aws_subnet.apps_b.cidr_block]
  }
}

resource "aws_security_group" "apps_allow_chat_public" {
  vpc_id = aws_vpc.main.id

  ingress {
    from_port   = local.chat_port
    to_port     = local.chat_port
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "apps_permissive_egress" {
  vpc_id = aws_vpc.main.id
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
  vpc_security_group_ids      = [aws_security_group.apps_allow_ssh.id, aws_security_group.apps_permissive_egress.id, aws_security_group.apps_allow_chat_local.id]
  availability_zone           = local.az
  associate_public_ip_address = true
  subnet_id                   = aws_subnet.apps_a.id
  user_data                   = local.init_intance
}

resource "aws_instance" "app_b" {
  ami                         = "ami-00051469f31042765"
  instance_type               = "t2.micro"
  key_name                    = aws_key_pair.default.key_name
  vpc_security_group_ids      = [aws_security_group.apps_allow_ssh.id, aws_security_group.apps_permissive_egress.id, aws_security_group.apps_allow_chat_local.id]
  availability_zone           = local.az
  associate_public_ip_address = true
  subnet_id                   = aws_subnet.apps_b.id
  user_data                   = local.init_intance
}

resource "aws_instance" "app_c" {
  ami                         = "ami-00051469f31042765"
  instance_type               = "t2.micro"
  key_name                    = aws_key_pair.default.key_name
  vpc_security_group_ids      = [aws_security_group.apps_allow_ssh.id, aws_security_group.apps_permissive_egress.id, aws_security_group.apps_allow_chat_public.id]
  availability_zone           = local.az
  associate_public_ip_address = true
  subnet_id                   = aws_subnet.apps_c.id
  user_data                   = local.init_intance
}

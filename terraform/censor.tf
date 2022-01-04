resource "aws_security_group" "censor_allow_geneve" {
  vpc_id = aws_vpc.main.id
  ingress {
    from_port   = local.geneve_port
    to_port     = local.geneve_port
    protocol    = "udp"
    cidr_blocks = [aws_subnet.appliances.cidr_block]
  }
}

resource "aws_security_group" "censor_allow_ssh" {
  vpc_id = aws_vpc.main.id
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "censor_allow_health_check" {
  vpc_id = aws_vpc.main.id
  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = [aws_subnet.appliances.cidr_block]
  }
}

resource "aws_security_group" "censor_permissive_egress" {
  vpc_id = aws_vpc.main.id
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}



resource "aws_instance" "appliance" {
  ami                         = "ami-00051469f31042765"
  instance_type               = "t2.micro"
  key_name                    = aws_key_pair.default.key_name
  vpc_security_group_ids      = [aws_security_group.censor_allow_ssh.id, aws_security_group.censor_allow_geneve.id, aws_security_group.censor_permissive_egress.id, aws_security_group.censor_allow_health_check.id]
  availability_zone           = local.az
  associate_public_ip_address = true
  subnet_id                   = aws_subnet.appliances.id
  user_data                   = local.init_intance
}


resource "aws_lb" "gateway" {
  load_balancer_type = "gateway"
  subnets            = [aws_subnet.appliances.id]
}


resource "aws_lb_target_group" "appliances" {
  port     = local.geneve_port
  protocol = "GENEVE"
  vpc_id   = aws_vpc.main.id
  health_check {
    enabled  = true
    protocol = "TCP"
    port     = 8080
    interval = 10
  }
}

resource "aws_lb_target_group_attachment" "appliances" {
  target_group_arn = aws_lb_target_group.appliances.arn
  target_id        = aws_instance.appliance.id
  port             = local.geneve_port
}
resource "aws_lb_listener" "gateway" {
  load_balancer_arn = aws_lb.gateway.arn
  default_action {
    target_group_arn = aws_lb_target_group.appliances.id
    type             = "forward"
  }
}

resource "aws_vpc_endpoint_service" "censor_service" {
  acceptance_required        = false
  allowed_principals         = [data.aws_caller_identity.current.arn]
  gateway_load_balancer_arns = [aws_lb.gateway.arn]
}

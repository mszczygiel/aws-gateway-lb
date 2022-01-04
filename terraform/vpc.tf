resource "aws_vpc" "main" {
  cidr_block = "192.168.0.0/16"
}
# resource "aws_s3_bucket" "main_flow_logs" {
#     bucket = "mszczygiel-demo-apps-flow-logs"
#     acl = "log-delivery-write"
#     force_destroy = true
# }

# resource "aws_flow_log" "main" {
#   log_destination      = aws_s3_bucket.main_flow_logs.arn
#   log_destination_type = "s3"
#   traffic_type         = "ALL"
#   vpc_id               = aws_vpc.main.id
# }


resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
}
resource "aws_route_table" "igw" {
  vpc_id = aws_vpc.main.id
}

resource "aws_route_table_association" "igw" {
  gateway_id     = aws_internet_gateway.main.id
  route_table_id = aws_route_table.igw.id
}

resource "aws_subnet" "apps_a" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "192.168.1.0/24"
  availability_zone = local.az
}
resource "aws_route_table" "apps_a" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block      = aws_subnet.apps_b.cidr_block
    vpc_endpoint_id = aws_vpc_endpoint.censored_communication.id
  }
  route {
    cidr_block      = "0.0.0.0/0"
    vpc_endpoint_id = aws_internet_gateway.main.id
  }
}

resource "aws_route_table_association" "apps_a" {
  subnet_id      = aws_subnet.apps_a.id
  route_table_id = aws_route_table.apps_a.id
}

resource "aws_subnet" "apps_b" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "192.168.2.0/24"
  availability_zone = local.az
}

resource "aws_route_table" "apps_b" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block      = aws_subnet.apps_a.cidr_block
    vpc_endpoint_id = aws_vpc_endpoint.censored_communication.id
  }
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
}

resource "aws_route_table_association" "apps_b" {
  subnet_id      = aws_subnet.apps_b.id
  route_table_id = aws_route_table.apps_b.id
}

resource "aws_subnet" "endpoint" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "192.168.10.0/24"
  availability_zone = local.az
}

resource "aws_route_table" "endpoint" {
  vpc_id = aws_vpc.main.id
}

resource "aws_route_table_association" "endpoint" {
  subnet_id      = aws_subnet.endpoint.id
  route_table_id = aws_route_table.endpoint.id
}

resource "aws_subnet" "appliances" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "192.168.20.0/24"
  availability_zone = local.az
}

resource "aws_route_table" "appliance" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
}

resource "aws_route_table_association" "appliance" {
  subnet_id      = aws_subnet.appliances.id
  route_table_id = aws_route_table.appliance.id
}

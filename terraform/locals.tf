locals {
  az           = "eu-central-1a"
  geneve_port  = 6081
  init_intance = <<EOF
#!/bin/bash
sleep 60
yum -y update && yum -y install nc > logs.txt
echo "TERM=vt100" >> /etc/environment
  EOF
  chat_port    = 3000
}

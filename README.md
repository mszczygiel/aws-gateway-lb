# Demo of AWS Gateway Load Balancer
This repository is a simple demo of AWS Gateway Load Balancer.

If packet contains string "drop me" then it will be dropped,

String "weakly typed" will be replaced with "strongly typed"

Other packets are forwarded without change

**Note that resources provisioned with this demo will incur some costs. Remember to destroy provisioned infrastructure.**
## Infrastructure
Infrastructure is managed by Terraform (`terraform` directory). It consists of:
- VPC (192.168.0.0/16)
- subnets
  - 192.168.1.0/24 -> app A
  - 192.168.2.0/24 -> app B
  - 192.168.3.0/24 -> app C
  - 192.168.10.0/24 -> local VPC endpoint
  - 192.168.100.0/24 -> VPC endpoint with routes from/to Internet Gateway
  - 192.168.20.0/24 -> Gateway Load Balancer + virtual appliance
- routing configuration
  - app A <--> app B traffic is routed through local VPC endpoint for inspection (public traffic without inspection)
  - app C traffic is routed through public VPC endpoint. **Note, that in order to SSH into this instance, censor application must run**
- "application" instances
  - A, B, C - will be used to launch netcat
- internet gateway
- VPC endpoint without Internet Access
- VPC endpoint with Internet access
- Gateway Load Balancer
- appliance instance - application that inspects packets. Target for Gateway Load Balancer
- security groups:
  - allowed ingress on port 22 on all instances (SSH)
  - allowed UDP traffic on port 3000 on all instances (for demo purposes)
  - allowed UDP traffic on port 6081 (GENEVE) on virtual appliance
  - allowed TCP traffic on port 8080 (health check) on virtual appliance
  - permissive egress on all instances 

## Virtual appliance
Source code of the virtual appliance can be found in the `censor` directory. The application captures raw packets and inspects GENEVE traffic. Packets are modified (replace "weakly typed" with "strongly typed"), dropped (if payload contains "drop me") or forwarded unmodified (in other cases)


## Prerequisites
- [Terraform](https://www.terraform.io/) and [tfswitch](https://tfswitch.warrensbox.com/)
- AWS access configured
- [go](https://go.dev) (version at least 1.17)
- public SSH key under `~/.ssh/id_rsa.pub`
- environment variable `TF_STATE_BUCKET` with name of the S3 bucket where terraform state will be stored

## How to run
- `./deploy_infra.sh` - provisions AWS infrastructure with terraform
- `./deploy_censor.sh` - builds, deploys and runs our virtual appliance
- `./init_infra.sh` - installs netcat on all apps (to install on app C, virtual appliance must run)

To see how local traffic is routed through the censor open 2 terminals:
- `./ssh.sh a` and run `nc -l -u -p 3000`
- `./ssh.sh b` and run `nc -u PRIVATE_IP_OF_APP_A 3000` then write sample messages like `hello`, `drop me`, `weakly typed`

To see how public traffic is routed through censor open 2 terminals:
- `./ssh.sh c` and run `nc -l -u -p 3000`
- `./connect.sh` and write sample messages

## Destroy infrastructure
To avoid unnecessary costs, remember do destroy provisioned infrastructure with `./destroy_infra.sh`

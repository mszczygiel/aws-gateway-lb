# Demo of AWS Gateway Load Balancer
This repository is a simple demo of AWS Gateway Load Balancer.

It deploys virtual appliance that captures Geneve packets and processes all that contain
- UDP packets where source or destination port is 3000
- ICMP packets

See [Virtual appliance section](#virtual-appliance) to learn more about how captured packets are handled.

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
  - app A <--> app B traffic is routed through local VPC endpoint and is inspected by the virtual appliance (public traffic goes without inspection)
  - app C traffic is routed through public VPC endpoint and is inspected by the virtual appliance. **Note, that in order to SSH into this instance, censor application must run** - this is because all public traffic (including SSH) goes through the virtual appliance.
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
  - allowed UDP traffic on port 6081 (Geneve) on virtual appliance
  - allowed TCP traffic on port 8080 (health check) on virtual appliance
  - allowed ICMP in 192.168.0.0/16 subnet
  - permissive egress on all instances 

## Virtual appliance
Source code of the virtual appliance can be found in the `censor` directory.

For UDP packets, if payload contains string "weakly typed" it's replaced with string "strongly typed". Additionally, if payload contains "drop me" string the whole packet is dropped.

For ICMP packets, every 5th one is dropped.

Other Geneve traffic is forwarded without changing packages' contents.

## Prerequisites
- computer running Linux 64-bit x86 architecture
- [Terraform](https://www.terraform.io/) and [tfswitch](https://tfswitch.warrensbox.com/)
- AWS access configured
- [go](https://go.dev) (version at least 1.20). Virtual appliance is implemented in the Go language.
- `jq` must be installed. It's used to extract instances' IP addresses from the TF state.
- public SSH key under `~/.ssh/id_rsa.pub`. The key is added to the deployed instances so you can ssh into them later.

## How to run
- run `./deploy_infra.sh` which provisions AWS infrastructure with terraform
- then run `./deploy_censor.sh` - builds, deploys and runs the appliance
- then run `./init_infra.sh` - installs netcat on all apps (to install on app C, virtual appliance must run)

To see how internal traffic is routed through the censor open 2 terminals:
- `./ssh.sh a` and run `nc -l -u -p 3000`
- `./ssh.sh b` and run `nc -u PRIVATE_IP_OF_APP_A 3000` then write sample messages like `hello`, `drop me`, `weakly typed`
- from one of the instances (either a or b) ping private IP of the other one

To see how public traffic is routed through censor open 2 terminals:
- `./ssh.sh c` and run `nc -l -u -p 3000`
- `./connect.sh` and write sample messages

## Destroy infrastructure
To avoid unnecessary costs, remember do destroy provisioned infrastructure with `./destroy_infra.sh`.

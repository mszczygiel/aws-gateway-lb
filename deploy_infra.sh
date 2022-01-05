#!/bin/sh -e
(cd terraform && tfswitch)
TF_REGISTRY_CLIENT_TIMEOUT=60 terraform -chdir=terraform init && \
terraform -chdir=terraform plan -refresh=true -out=plan.tfplan && \
terraform -chdir=terraform apply plan.tfplan

state=$(terraform -chdir=terraform show -json -no-color)
pub_ip_a=$(echo $state | jq -c '.values.root_module.resources[] | select(.address == "aws_instance.app_a") | .values.public_ip')
pub_ip_b=$(echo $state | jq -c '.values.root_module.resources[] | select(.address == "aws_instance.app_b") | .values.public_ip')
pub_ip_c=$(echo $state | jq -c '.values.root_module.resources[] | select(.address == "aws_instance.app_c") | .values.public_ip')
pub_ip_appliance=$(echo $state | jq -c '.values.root_module.resources[] | select(.address == "aws_instance.appliance") | .values.public_ip')

echo "export APP_A_IP=$pub_ip_a" > IPS.sh
echo "export APP_B_IP=$pub_ip_b" >> IPS.sh
echo "export APP_C_IP=$pub_ip_c" >> IPS.sh
echo "export APPLIANCE_IP=$pub_ip_appliance" >> IPS.sh

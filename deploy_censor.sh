#!/bin/sh -e

source ./IPS.sh

(cd censor && CGO_ENABLED=0 go build)

ssh -o StrictHostKeyChecking=no ec2-user@$APPLIANCE_IP sh -c '"sudo pkill censor"' || true
scp  ./censor/censor ec2-user@$APPLIANCE_IP:/home/ec2-user/censor
ssh ec2-user@$APPLIANCE_IP sh -c '"sudo ./censor"'

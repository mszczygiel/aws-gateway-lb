#!/bin/sh

source ./IPS.sh

(cd censor && CGO_ENABLED=0 go build)
scp ./censor/censor ec2-user@$APPLIANCE_IP:/home/ec2-user/censor

#!/bin/sh

source ./IPS.sh

scp ./censor/censor ec2-user@$APPLIANCE_IP:/home/ec2-user/censor

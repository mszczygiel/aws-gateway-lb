#!/bin/sh

source ./IPS.sh

CMD=$1

case $CMD in 
  "a")
    IP=$APP_A_IP ;;
  "b")
    IP=$APP_B_IP ;;
  "appliance")
    IP=$APPLIANCE_IP ;;
esac

ssh ec2-user@$IP

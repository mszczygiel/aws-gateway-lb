#!/bin/sh -e

. ./IPS.sh

CMD=$1

case $CMD in 
  "a")
    IP=$APP_A_IP ;;
  "b")
    IP=$APP_B_IP ;;
  "c")
    IP=$APP_C_IP ;;
  "appliance")
    IP=$APPLIANCE_IP ;;
esac

ssh -o StrictHostKeyChecking=no ec2-user@$IP

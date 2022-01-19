#!/bin/sh -e

source ./IPS.sh

ssh -o StrictHostKeyChecking=no ec2-user@$APP_A_IP sh -c '"sudo yum -y install nc"'
ssh -o StrictHostKeyChecking=no ec2-user@$APP_B_IP sh -c '"sudo yum -y install nc"'
ssh -o StrictHostKeyChecking=no ec2-user@$APP_C_IP sh -c '"sudo yum -y install nc"'

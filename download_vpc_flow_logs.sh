#!/bin/sh -e

APPS_DIR="logs/vpc/app"
CENSOR_DIR="logs/vpc/censor"

mkdir -p $APPS_DIR
mkdir -p $CENSOR_DIR

aws s3 sync s3://mszczygiel-demo-apps-flow-logs $APPS_DIR
aws s3 sync s3://mszczygiel-demo-censor-flow-logs $CENSOR_DIR

(cd $APPS_DIR && gunzip -f -r .)
(cd $CENSOR_DIR && gunzip -f -r .)

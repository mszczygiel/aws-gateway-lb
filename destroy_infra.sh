#!/bin/sh -e

(cd terraform && tfswitch)
terraform -chdir=terraform destroy -auto-approve

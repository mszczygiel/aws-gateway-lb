#!/bin/sh

(cd terraform && tfswitch)
terraform -chdir=terraform destroy -auto-approve

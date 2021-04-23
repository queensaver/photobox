#!/bin/bash

set -e
set -o pipefail

sudo apt update
sudo apt install -y ansible git

if [ -e bbox ]; then
  git -C photobox/ pull
else
  git clone https://github.com/queensaver/photobox.git
fi

ansible-playbook photobox/ansible/photobox.yml
ansible-playbook photobox/ansible/read_only.yml

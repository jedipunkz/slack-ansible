# Slack-Ansible

Slack Bot which deploys server by ansible command, and output logs to aws s3.

<img src="https://raw.githubusercontent.com/jedipunkz/slack-ansible/master/pix/slack-ansible.png">

## Getting Started

### Prerequisites

What things you need to install this software and how to install them.

* Linux OS or Apple macOS
* Golang 1.13.x or later
* ansible
* aws credentials (now, this software applied to use [localstack](https://github.com/localstack/localstack)

### Setup $HOME/.slack-ansible.yaml file

```bash
cat << EOF > $HOME/.slack-ansible.yaml
token: <your bot's token>
EOF
```

### Installation

```bash
go get github.com/jedipunkz/slack-ansible
```

### boot

```bash
nohup /path/to/slack-ansible &
```

## Author

Tomokazu HIRAI <tomokazu.hirai@gmail.com>

## License

This project is licensed under the Apache License.

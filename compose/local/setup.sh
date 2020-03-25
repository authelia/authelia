#!/bin/bash

username(){
  read -ep "Enter your username for Authelia: " USERNAME
}

password(){
  read -esp "Enter a password for $USERNAME: " PASSWORD
}

echo "Checking for pre-requisites"

if [[ ! -x "$(command -v docker)" ]]; then
  echo "You must install Docker on your machine";
  return
fi

if [[ ! -x "$(command -v docker-compose)" ]]; then
  echo "You must install Docker Compose on your machine";
  return
fi

echo "Pulling Authelia docker image for setup"
docker pull authelia/authelia > /dev/null

read -ep "What root domain would you like to protect? (default/no selection is example.com): " DOMAIN

if [[ $DOMAIN == "" ]]; then
  DOMAIN="example.com"
fi

MODIFIED=$(cat /etc/hosts | grep $DOMAIN && echo true || echo false)

if [[ $MODIFIED == "false" ]]; then
echo "\
127.0.0.1  authelia.$DOMAIN
127.0.0.1  public.$DOMAIN
127.0.0.1  traefik.$DOMAIN
127.0.0.1  secure.$DOMAIN" >> /etc/hosts
fi

echo "Generating SSL certificate for *.$DOMAIN"
docker run -a stdout -v $PWD/traefik/certs:/tmp/certs authelia/authelia authelia certificates generate --host *.$DOMAIN --dir /tmp/certs/ > /dev/null

if [[ $DOMAIN != "example.com" ]]; then
  sed -i "s/example.com/$DOMAIN/g" {docker-compose.yml,configuration.yml}
fi

username

if [[ $USERNAME != "" ]]; then
  sed -i "s/<USERNAME>/$USERNAME/g" users_database.yml
else
  echo "Username cannot be empty"
  username
fi

password

if [[ $PASSWORD != "" ]]; then
  PASSWORD=$(docker run authelia/authelia authelia hash-password $PASSWORD | sed 's/Password hash: //g')
  sed -i "s/<PASSWORD>/$(echo $PASSWORD | sed -e 's/[\/&]/\\&/g')/g" users_database.yml
else
  echo "Password cannot be empty"
  password
fi

cat << EOF
Setup completed successfully, please start up containers with 'docker-compose up -d'.

Once containers have been started you can now visit the following locations:
- https://public.$DOMAIN - Bypasses Authelia
- https://traefik.$DOMAIN - Secured with Authelia one-factor authentication
- https://secure.$DOMAIN - Secured with Authelia two-factor authentication

Once you have registered an OTP device, the link to generate your QR code will be in 'compose/local/authelia/notifications.txt'.
'grep "<a href=" compose/local/authelia/notifications.txt'
EOF


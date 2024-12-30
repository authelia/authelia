#!/usr/bin/env bash

writehosts(){
  echo "\
127.0.0.1  authelia.$DOMAIN
127.0.0.1  public.$DOMAIN
127.0.0.1  traefik.$DOMAIN
127.0.0.1  secure.$DOMAIN" | sudo tee -a /etc/hosts > /dev/null
}

username(){
  read -ep "Enter your username for Authelia: " USERNAME
}

password(){
  read -esp "Enter a password for $USERNAME: " PASSWORD
}

displayname(){
  read -ep "Enter your display name for Authelia (eg. John Doe): " DISPLAYNAME
}

echo "Checking for pre-requisites"

if [[ ! -x "$(command -v docker)" ]]; then
  echo "You must install Docker on your machine";
  exit 1
fi

if [[ ! -x "$(command -v docker-compose)" ]]; then
  echo "You must install Docker Compose on your machine";
  exit 1
fi

if [[ $(id -u)  != 0 ]]; then
  echo "The script requires root access to perform some functions such as modifying your /etc/hosts file"
  read -rp "Would you like to elevate access with sudo? [y/N] " confirmsudo
  if ! [[ "$confirmsudo" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo "Sudo elevation denied, exiting"
    exit 1
  fi
fi

echo "Pulling Authelia docker image for setup"
sudo docker pull authelia/authelia > /dev/null

echo "Resetting docker-compose.yml, configuration.yml and users_database.yml"
sudo git checkout -- docker-compose.yml authelia/configuration.yml authelia/users_database.yml

read -ep "What root domain would you like to protect? (default/no selection is example.com): " DOMAIN

if [[ $DOMAIN == "" ]]; then
  DOMAIN="example.com"
fi

MODIFIED=$(cat /etc/hosts | grep $DOMAIN && echo true || echo false)

if [[ $MODIFIED == "false" ]]; then
  writehosts
fi

echo "Generating SSL certificate for *.$DOMAIN"
sudo docker run -a stdout -v $PWD/traefik/certs:/tmp/certs authelia/authelia authelia crypto certificate rsa generate --common-name="*.${DOMAIN}" --directory=/tmp/certs/ > /dev/null

if [[ $DOMAIN != "example.com" ]]; then
  if [[ $(uname) == "Darwin" ]]; then
    sudo sed -i '' "s/example.com/$DOMAIN/g" {docker-compose.yml,authelia/configuration.yml}
  else
    sudo sed -i "s/example.com/$DOMAIN/g" {docker-compose.yml,authelia/configuration.yml}
  fi
fi

username

if [[ $USERNAME != "" ]]; then
  if [[ $(uname) == "Darwin" ]]; then
    sudo sed -i '' "s/<USERNAME>/$USERNAME/g" authelia/users_database.yml
  else
    sudo sed -i "s/<USERNAME>/$USERNAME/g" authelia/users_database.yml
  fi
else
  echo "Username cannot be empty"
  username
fi

displayname

if [[ $DISPLAYNAME != "" ]]; then
  if [[ $(uname) == "Darwin" ]]; then
    sudo sed -i '' "s/<DISPLAYNAME>/$DISPLAYNAME/g" authelia/users_database.yml
  else
    sudo sed -i "s/<DISPLAYNAME>/$DISPLAYNAME/g" authelia/users_database.yml
  fi
else
  echo "Display name cannot be empty"
  displayname
fi

password

if [[ $PASSWORD != "" ]]; then
  PASSWORD=$(sudo docker run authelia/authelia authelia crypto hash generate argon2 --password $PASSWORD | sed 's/Digest: //g')
  if [[ $(uname) == "Darwin" ]]; then
    sudo sed -i '' "s/<PASSWORD>/$(echo $PASSWORD | sed -e 's/[\/&]/\\&/g')/g" authelia/users_database.yml
  else
    sudo sed -i "s/<PASSWORD>/$(echo $PASSWORD | sed -e 's/[\/&]/\\&/g')/g" authelia/users_database.yml
  fi
else
  echo "Password cannot be empty"
  password
fi

sudo docker compose up -d

if [[ $? != 0 ]]; then
  exit 1
fi

cat << EOF
Setup completed successfully.

You can now visit the following locations:
- https://public.$DOMAIN - Bypasses Authelia
- https://traefik.$DOMAIN - Secured with Authelia one-factor authentication
- https://secure.$DOMAIN - Secured with Authelia two-factor authentication (see note below)

You will need to authorize the self-signed certificate upon visiting each domain.
To visit https://secure.$DOMAIN you will need to register a device for second factor authentication and confirm by clicking on a link sent by email. Since this is a demo with a fake email address, the content of the email will be stored in './authelia/notification.txt'.
Upon registering, you can grab this link easily by running the following command: 'grep -Eo '"https://.*" ' ./authelia/notification.txt'.
EOF

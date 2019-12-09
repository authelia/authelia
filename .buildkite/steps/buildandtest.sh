#!/bin/bash
set -eo pipefail

echo "--- :gear: Installing pre-requisites"

apt --fix-broken install -y
apt update -y
apt install build-essential fonts-liberation libappindicator3-1 libasound2 libatk-bridge2.0-0 libatk1.0-0 libatspi2.0-0 libcairo2 libcups2 libdbus-1-3 libgdk-pixbuf2.0-0 libgif-dev libglib2.0-0 libgtk-3-0 libnspr4 libnss3 libpango-1.0-0 libpangocairo-1.0-0 libxcomposite1 libxcursor1 libxi6 libxrandr2 libxrender1 libxss1 libxtst6 lsb-release unzip wget xdg-utils xvfb -y -q

if [ ! -f "/usr/bin/google-chrome-stable" ]; then
  cd /tmp
  curl -o google-chrome-stable_current_amd64.deb https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
  dpkg -i google-chrome-stable_current_amd64.deb
fi

if [ ! -d "/usr/local/go" ]; then
  cd /tmp
  curl -o go1.13.5.linux-amd64.tar.gz https://dl.google.com/go/go1.13.5.linux-amd64.tar.gz
  tar -xf go1.13.5.linux-amd64.tar.gz
  mv go /usr/local
fi

if [ ! -d "$HOME/.nvm" ]; then
  curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.35.1/install.sh | bash
  [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
  nvm install v12 && nvm use v12
fi

if [ ! -f "/usr/local/bin/chromedriver" ]; then
  cd /tmp
  curl -o chromedriver_linux64.zip https://chromedriver.storage.googleapis.com/78.0.3904.70/chromedriver_linux64.zip
  unzip chromedriver_linux64.zip -d ./
  rm chromedriver_linux64.zip
  mv -f chromedriver /usr/local/bin/chromedriver
  chmod +x /usr/local/bin/chromedriver
fi

echo "--- :hammer: Build and Test ${BUILDKITE_PIPELINE_SLUG}"
cd $BUILDKITE_BUILD_CHECKOUT_PATH
authelia-scripts --log-level debug ci

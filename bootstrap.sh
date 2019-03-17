
export PATH=$(pwd)/scripts:/tmp:$PATH

export PS1="(authelia) $PS1"

echo "[BOOTSTRAP] Installing npm packages..."
npm i

pushd client
npm i
popd

echo "[BOOTSTRAP] Checking if Docker is installed..."
if [ ! -x "$(command -v docker)" ];
then
  echo "[ERROR] You must install docker on your machine.";
  return
fi

echo "[BOOTSTRAP] Checking if docker-compose is installed..."
if [ ! -x "$(command -v docker-compose)" ];
then
  echo "[ERROR] You must install docker-compose on your machine.";
  return;
fi

echo "[BOOTSTRAP] Checking if example.com domain is forwarded to your machine..."
cat /etc/hosts | grep "login.example.com" > /dev/null
if [ $? -ne 0 ];
then
  echo "[ERROR] Please add those lines to /etc/hosts:
  
127.0.0.1       home.example.com
127.0.0.1       public.example.com
127.0.0.1       secure.example.com
127.0.0.1       dev.example.com
127.0.0.1       admin.example.com
127.0.0.1       mx1.mail.example.com
127.0.0.1       mx2.mail.example.com
127.0.0.1       singlefactor.example.com
127.0.0.1       login.example.com"
  return;
fi

echo "[BOOTSTRAP] Running additional bootstrap steps..."
authelia-scripts bootstrap

echo "[BOOTSTRAP] Run 'authelia-scripts suites start dockerhub' to start Authelia and visit https://home.example.com:8080."
echo "[BOOTSTRAP] More details at https://github.com/clems4ever/authelia/blob/master/docs/getting-started.md"

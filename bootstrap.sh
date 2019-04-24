
export PATH=$(pwd)/scripts:/tmp:$PATH

export PS1="(authelia) $PS1"

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

echo "[BOOTSTRAP] Running additional bootstrap steps..."
authelia-scripts bootstrap

# Create temporary directory that will contain the databases used in tests.
mkdir -p /tmp/authelia

echo "[BOOTSTRAP] Run 'authelia-scripts suites start dockerhub' to start Authelia and visit https://home.example.com:8080."
echo "[BOOTSTRAP] More details at https://github.com/clems4ever/authelia/blob/master/docs/getting-started.md"

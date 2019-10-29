
set -e

export PATH=./cmd/authelia-scripts/:/tmp:$PATH

if [ -z "$OLD_PS1" ]; then
  OLD_PS1="$PS1"
  export PS1="(authelia) $PS1"
fi


echo "[BOOTSTRAP] Checking if Go is installed..."
if [ ! -x "$(command -v go)" ];
then
  echo "[ERROR] You must install Go on your machine.";
  return
fi

authelia-scripts bootstrap

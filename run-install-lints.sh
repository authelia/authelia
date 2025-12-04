sudo mkdir -p ~/.local/bin && \
sudo curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin && \
curl -sSfL https://raw.githubusercontent.com/trufflesecurity/trufflehog/main/scripts/install.sh | sh -s -- -b ~/.local/bin
sudo apt update && sudo apt install -y yamllint && \
echo 'export PATH=$PATH:'"$(go env GOPATH)/bin"':~/.local/bin' >> ~/.bashrc
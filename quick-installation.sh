#!/bin/bash

clear
echo -e "\033[1;32m
██████╗ ██╗   ██╗██████╗ ██████╗ ██████╗ ██████╗  █████╗ 
██╔══██╗╚██╗ ██╔╝██╔══██╗██╔══██╗██╔══██╗╚════██╗██╔══██╗
██████╔╝ ╚████╔╝ ██║  ██║██║  ██║██║  ██║ █████╔╝╚██████║
██╔══██╗  ╚██╔╝  ██║  ██║██║  ██║██║  ██║██╔═══╝  ╚═══██║
██║  ██║   ██║   ██████╔╝██████╔╝██████╔╝███████╗ █████╔╝
╚═╝  ╚═╝   ╚═╝   ╚═════╝ ╚═════╝ ╚═════╝ ╚══════╝ ╚════╝ 
\033[0m"
echo -e "\033[1;34m==================================================\033[0m"
echo -e "\033[1;34m@Ryddd29 | Testnet, Node Runner, Developer, Retrodrop\033[0m"

sleep 4

echo ''
# Update package
echo -e "\033[1;32m\033[1mUpdating & Upgrading packages...\033[0m"
sudo apt update && sudo apt upgrade -y
sudo apt install git -y
clear

# Prompt to ask user if they want to install Go version 1.23.1
read -p $'\033[1;32m\033[1mDo you want to install Go version 1.23.1? (y/n) [default: y]: \033[0m' USER_INPUT

# Default to "y" if no input is provided
USER_INPUT=${USER_INPUT:-y}

if [[ "$USER_INPUT" =~ ^[Yy]$ ]]; then
  echo -e "\033[1;32m\033[1mInstalling Go version 1.23.1...\033[0m"

  # Define the URL for Go 1.23.1 (64-bit Linux)
  GO_URL="https://go.dev/dl/go1.23.1.linux-amd64.tar.gz"

  # Download and install Go
  echo -e "\033[1;32mDownloading Go from: $GO_URL\033[0m"
  curl -LO $GO_URL
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf "go1.23.1.linux-amd64.tar.gz"
  rm "go1.23.1.linux-amd64.tar.gz"

  # Add Go to the PATH
  echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
  source ~/.bashrc

  # Verify installation
  if go version &>/dev/null; then
    echo -e "\033[0;32mGo $(go version) successfully installed.\033[0m"
  else
    echo -e "\033[0;31mFailed to install Go.\033[0m"
  fi
else
  echo -e "\033[0;33mInstallation skipped by user.\033[0m"
fi

# Clone GitHub repository
echo -e "\033[1;32m\033[1mCloning GitHub repository...\033[0m"
git clone https://github.com/initia-labs/weave.git
cd weave

# Install dependencies
echo -e "\033[1;32m\033[1mInstalling required dependencies...\033[0m"
git checkout tags/v0.1.1
make install

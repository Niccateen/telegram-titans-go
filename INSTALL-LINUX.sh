#!/bin/bash

echo "================================"
echo "TDL GUI INSTALLER FOR LINUX"
echo "================================"
echo ""

# Check if running as root
if [ "$EUID" -eq 0 ]; then
    echo "dont run this as root, run it as normal user"
    exit 1
fi

echo "STEP 1 - Installing tdl..."
curl -L -o tdl.tar.gz https://github.com/iyear/tdl/releases/latest/download/tdl_Linux_64bit.tar.gz
tar -xzf tdl.tar.gz
sudo mv tdl /usr/local/bin/
rm tdl.tar.gz
echo "tdl installed"
echo ""

echo "STEP 2 - Installing Go 1.23..."
sudo rm -rf /usr/local/go
wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
rm go1.23.4.linux-amd64.tar.gz
echo "go installed"
echo ""

echo "STEP 3 - Installing GUI dependencies..."
sudo apt-get update
sudo apt-get install -y libgl1-mesa-dev xorg-dev libxxf86vm-dev
echo "dependencies installed"
echo ""

echo "STEP 4 - Building GUI..."
go mod tidy
go build -o tdl-gui .
echo "gui built"
echo ""

echo "STEP 5 - Installing GUI..."
sudo mv tdl-gui /usr/local/bin/
sudo cp tdl-gui.desktop /usr/share/applications/
echo "gui installed"
echo ""

echo "================================"
echo "INSTALLATION COMPLETE"
echo "================================"
echo ""
echo "NOW YOU NEED TO LOGIN TO TELEGRAM"
echo "run this command:"
echo ""
echo "    tdl login"
echo ""
echo "after login you can open the app by:"
echo "- typing tdl-gui in terminal"
echo "- or find TDL GUI in your applications menu"
echo ""
echo "press enter to exit"
read


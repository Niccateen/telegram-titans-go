#!/bin/bash

echo "================================"
echo "TDL GUI INSTALLER FOR MAC"
echo "================================"
echo ""

# Get the directory where this script is located
cd "$(dirname "$0")"

echo "STEP 1 - Checking Homebrew..."
if ! command -v brew &> /dev/null; then
    echo "Homebrew not installed. Installing..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
fi
echo "Homebrew ready"
echo ""

echo "STEP 2 - Installing Go..."
brew install go
echo "Go installed"
echo ""

echo "STEP 3 - Installing tdl..."
brew install iyear/tap/tdl
echo "tdl installed"
echo ""

echo "STEP 4 - Building GUI..."
go mod tidy
go build -o tdl-gui .
echo "GUI built"
echo ""

echo "STEP 5 - Making it executable..."
chmod +x tdl-gui
echo ""

echo "================================"
echo "INSTALLATION COMPLETE"
echo "================================"
echo ""
echo "NOW YOU NEED TO LOGIN TO TELEGRAM"
echo "open Terminal and run:"
echo ""
echo "    tdl login"
echo ""
echo "after login you can open the app by:"
echo "- double clicking tdl-gui in this folder"
echo "- or running ./tdl-gui in terminal"
echo ""
echo "press enter to exit"
read


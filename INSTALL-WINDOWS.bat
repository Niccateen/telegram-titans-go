@echo off
echo ================================
echo TDL GUI INSTALLER FOR WINDOWS
echo ================================
echo.

echo STEP 1 - Checking Go...
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo Go is not installed
    echo Download Go from https://go.dev/dl/
    echo Install it then run this script again
    pause
    exit /b 1
)
echo Go is installed
echo.

echo STEP 2 - Checking tdl...
tdl version >nul 2>&1
if %errorlevel% neq 0 (
    echo tdl is not installed
    echo Download tdl from https://github.com/iyear/tdl/releases
    echo Get the Windows zip, extract tdl.exe somewhere in your PATH
    pause
    exit /b 1
)
echo tdl is installed
echo.

echo STEP 3 - Building GUI...
go mod tidy
go build -o tdl-gui.exe .
echo.

echo ================================
echo BUILD COMPLETE
echo ================================
echo.
echo The file tdl-gui.exe is now in this folder
echo.
echo NOW YOU NEED TO LOGIN TO TELEGRAM
echo Open Command Prompt and run:
echo.
echo     tdl login
echo.
echo After login double click tdl-gui.exe to open the app
echo.
pause


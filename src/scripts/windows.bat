@echo off
setlocal

REM Renk kodları Windows Komut İsteminde doğrudan çalışmaz, bu yüzden normal metin kullanıyoruz.
echo Starting 'huh' command line tool installation...

REM Go derleyicisinin varlığını kontrol et
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo Go compiler not found.
    echo Please install Go first and make sure it's in your PATH.
    echo You can download it from: https://golang.org/dl/
    exit /b 1
)

echo Go compiler found.

REM Proje kök dizinini bul (mevcut dizinin iki üstü)
pushd ..
pushd ..
set "PROJECT_ROOT_DIR=%CD%"
popd
popd

set "CURRENT_DIR=%CD%"

echo Going to the project root directory to handle Go modules: %PROJECT_ROOT_DIR%
cd /D "%PROJECT_ROOT_DIR%"
if %errorlevel% neq 0 (
    echo Failed to change to the project root directory.
    exit /b 1
)

REM go.mod dosyasının varlığını kontrol et
if not exist "go.mod" (
    echo "go.mod" file not found. Initializing Go module...
    go mod init huh-cli
    if %errorlevel% neq 0 (
        echo Failed to initialize Go module.
        exit /b 1
    )
    echo Go module initialized successfully.
)

echo Downloading and tidying Go dependencies...
go mod tidy
if %errorlevel% neq 0 (
    echo An error occurred while managing Go dependencies.
    exit /b 1
)
echo Go dependencies are ready.

echo Returning to the installation script directory: %CURRENT_DIR%
cd /D "%CURRENT_DIR%"
if %errorlevel% neq 0 (
    echo Failed to return to the installation script directory.
    exit /b 1
)

echo Compiling the 'huh' binary...
REM Windows için .exe uzantılı derleme
go build -o huh.exe ../main.go

REM Derlemenin başarılı olup olmadığını kontrol et
if not exist "huh.exe" (
    echo Compilation failed. Please check the errors above.
    exit /b 1
)

echo Compilation successful.

REM Kurulum için önerilen bir dizin oluşturalım (Örn: C:\Program Files\huh)
set "INSTALL_DIR=%ProgramFiles%\huh"
echo Attempting to install 'huh' to "%INSTALL_DIR%"...

if not exist "%INSTALL_DIR%" (
    mkdir "%INSTALL_DIR%"
)

REM Derlenmiş dosyayı kurulum dizinine taşı
move huh.exe "%INSTALL_DIR%\huh.exe"
if %errorlevel% neq 0 (
    echo Failed to move 'huh.exe' to "%INSTALL_DIR%".
    echo Please try running this script as an Administrator.
    exit /b 1
)

echo.
echo ====================================================================
echo      'huh' was successfully installed to "%INSTALL_DIR%"
echo ====================================================================
echo.
echo To use the 'huh' command anywhere in your terminal, you need to
echo add this directory to your system's PATH environment variable.
echo.
echo You can do this manually by following these steps:
echo   1. Search for 'Edit the system environment variables' in the Start Menu.
echo   2. Click the 'Environment Variables...' button.
echo   3. Under 'System variables', find and select the 'Path' variable, then click 'Edit'.
echo   4. Click 'New' and add the following path:
echo      %INSTALL_DIR%
echo   5. Click OK on all windows to save the changes.
echo.
echo After adding it to your PATH, please restart your terminal.
echo Try it out by typing: huh help
echo.

endlocal

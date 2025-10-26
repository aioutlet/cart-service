@echo off
:loop
echo.
echo ============================================
echo Starting cart service...
echo ============================================
echo.

REM Check if port 8085 is in use and kill the process
echo Checking port 8085...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8085 ^| findstr LISTENING') do (
    echo Port 8085 is in use by PID %%a, killing process...
    taskkill /PID %%a /F >nul 2>&1
    timeout /t 1 >nul
)

echo Starting service on port 8085...
go run ./cmd/server

echo.
echo ============================================
echo Service stopped. Press any key to restart or Ctrl+C to exit.
echo ============================================
pause > nul
goto loop

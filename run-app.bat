@echo off
REM Load environment variables from .env file
for /f "tokens=*" %%a in (.env) do (
    echo %%a | findstr /v "^#" > nul
    if not errorlevel 1 (
        for /f "tokens=1,2 delims==" %%b in ("%%a") do (
            set %%b=%%c
        )
    )
)

REM Run the application
cd backend
go run main.go

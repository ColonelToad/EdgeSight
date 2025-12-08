@echo off
echo Starting EdgeSight Dashboard...
echo.
echo Frontend will be available at: http://localhost:8000
echo API server should be running at: http://localhost:8080
echo.
echo Press Ctrl+C to stop the server
echo.

REM Check if Python is available
python --version >nul 2>&1
if %errorlevel% == 0 (
    echo Using Python to serve frontend...
    python -m http.server 8000
    goto :eof
)

REM Check if PHP is available
php --version >nul 2>&1
if %errorlevel% == 0 (
    echo Using PHP to serve frontend...
    php -S localhost:8000
    goto :eof
)

echo ERROR: Neither Python nor PHP found!
echo Please install Python or PHP to run the frontend.
echo Or manually open index.html in your browser with CORS disabled.
pause

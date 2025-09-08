@echo off
echo Generating protobuf files...
mkdir pb 2>nul
protoc --go_out=./ --go-grpc_out=./ order.proto
if %ERRORLEVEL% neq 0 (
    echo Error: Failed to generate protobuf files
    pause
    exit /b 1
)
echo Done!
pause
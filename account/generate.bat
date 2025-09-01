@echo off
echo Generating protobuf files...
protoc --go_out=./ --go-grpc_out=./ account.proto
echo Done!
pause

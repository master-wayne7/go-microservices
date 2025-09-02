@echo off
echo Generating protobuf files...
mkdir pb
protoc --go_out=./ --go-grpc_out=./ catalog.proto
echo Done!
pause

@echo off
echo GoCryptoTrader: Generating gRPC, proxy and swagger files.
REM You may need to include the go mod package for the annotations file:
REM %GOPATH%\go\pkg\mod\github.com\grpc-ecosystem\grpc-gateway\v2@v2.0.1\third_party\googleapis

protoc -I=. -I=%GOPATH%\src -I=%GOPATH%\src\github.com\grpc-ecosystem\grpc-gateway\third_party\googleapis --go_out=. rpc.proto
protoc -I=. -I=%GOPATH%\src -I=%GOPATH%\src\github.com\grpc-ecosystem\grpc-gateway\third_party\googleapis --go-grpc_out=. rpc.proto
protoc -I=. -I=%GOPATH%\src -I=%GOPATH%\src\github.com\grpc-ecosystem\grpc-gateway\third_party\googleapis --grpc-gateway_out=logtostderr=true:. rpc.proto
protoc -I=. -I=%GOPATH%\src -I=%GOPATH%\src\github.com\grpc-ecosystem\grpc-gateway\third_party\googleapis --swagger_out=logtostderr=true:. rpc.proto
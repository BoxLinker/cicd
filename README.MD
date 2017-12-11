gen grpc:
    protoc -I . --go_out=plugins=grpc:. ./boxci.proto
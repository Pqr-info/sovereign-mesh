#!/bin/bash
echo -e "\033[93m[PROTO] Compiling sync.proto...\033[0m"

# Ensure target directory is clean and regenerate
rm -f grpc_node/sync_pb2*.py

python3 -m grpc_tools.protoc \
    -Iproto \
    --python_out=grpc_node \
    --grpc_python_out=grpc_node \
    proto/sync.proto proto/mesh_proto.proto

echo -e "\033[92m[PROTO] Compilation successful!\033[0m"

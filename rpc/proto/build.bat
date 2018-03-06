protoc.exe --plugin=protoc-gen-go.exe="protoc-gen-go" --go_out=../../server/src/rpc/ *.proto
protoc.exe --cpp_out=./ *.proto



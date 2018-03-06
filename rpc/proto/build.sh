protoc --plugin=protoc-gen-go="protoc-gen-go" --go_out=../../server/src/rpc/ *.proto
protoc --cpp_out=./ *.proto



#make

projectpath=$(pwd)

export GOROOT=/usr/local/go
#export GOROOT=/Users/fengqiang/Downloads/go
export GOBIN=$GOROOT/bin
export GOPATH=$projectpath/server/3rdpkg:$projectpath/server:$GOROOT
export PATH=$PATH:$GOROOT/bin:$GOBIN

echo $(go version)
go clean

cd bin

go build tools/dbtool
echo build dbtool ok!
go build tools/cachetool
echo build dbtool ok!

#export GOPATH=$projectpath/server/tools/GmTools:$GOPATH
#go build -o ../server/tools/GmTools/gmtools/gmtools ../server/tools/GmTools/gmtools/main.go
#echo build gmtools ok!

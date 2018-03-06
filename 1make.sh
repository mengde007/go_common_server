projectpath=$(pwd)

export GOROOT=/usr/local/go
export GOBIN=$GOROOT/bin
export GOPATH=$projectpath/server/3rdpkg:$projectpath/server:$GOROOT
export PATH=$PATH:$GOROOT/bin:$GOBIN

echo $(go version)
go clean

svn up

sh make.sh

cd bin

sh 1killall.sh

sh open.sh

cd ..

projectpath=$(pwd)

export GOROOT=/usr/local/go
export GOBIN=$GOROOT/bin
export GOPATH=$projectpath/server/3rdpkg:$projectpath/server:$GOROOT
export PATH=$PATH:$GOROOT/bin:$GOBIN

echo $(go version)
go clean

cd bin
go build tools/payserver
echo build payserver ok!

go build tools/gameserver
echo build gameserver ok !

go build tools/dbserver
echo build dbserver ok !

go build tools/accountserver
echo build accountserver ok !

go build tools/centerserver
echo build centerserver ok !

go build tools/roleserver
echo build roleserver ok !

go build tools/daerserver
echo build daerserver ok !

go build tools/lockserver
echo build lockserver ok!

go build tools/roomserver
echo build roomserver ok!

go build tools/mailserver
echo build mailserver ok!

go build tools/gmserver
echo build gmserver ok!

go build tools/gmproxy
echo build gmproxy ok!

go build tools/matchdaerserver
echo build matchdaerserver ok!

go build tools/matchserver
echo build matchserver ok!

go build tools/pockerserver
echo build pockerserver ok!

go build tools/majiangserver
echo build majiangserver ok!

go build tools/rankserver
echo build rankserver ok!



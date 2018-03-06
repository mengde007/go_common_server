echo off
set projectpath=%~dp0

set GOROOT=E:\Go\
set GOBIN=%GOROOT%bin
set GOPATH=%projectpath%server\3rdpkg;%projectpath%server;%GOROOT%


go clean

cd bin

rem go build tools/gameserver
rem echo build gameserver ok !

rem go build tools/dbserver
rem echo build dbserver ok !

rem go build tools/accountserver
rem echo build accountserver ok !

rem go build tools/centerserver
rem echo build centerserver ok !

go build tools/daerserver
echo build daerserver ok !

cd ..




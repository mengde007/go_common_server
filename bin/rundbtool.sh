#!/bin/sh
./dbtool -o ../cfg/dbserverBase1.json -n ../cfg/dbserverBase1_new.json
##./dbtool -o ../cfg/dbserverExtern1.json -n ../cfg/dbserverExtern1_new.json
#./dbtool -o ../cfg/accountserver1.json -n ../cfg/accountserver1_new.json

./cachetool -o1 ../cfg/matchserver1.json -o2 ../cfg/matchserver2.json -o3 ../cfg/matchserver3.json -n ../cfg/matchserver4.json

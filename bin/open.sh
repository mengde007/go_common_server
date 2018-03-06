
echo "start dbbase.extern.. "
./dbserver -c ../cfg/dbserverBase1.json &
./dbserver -c ../cfg/dbserverExtern1.json &
echo "start account,lock..."
./accountserver -c ../cfg/accountserver1.json &
./lockserver -c ../cfg/lockserver1.json &

./rankserver &
echo "star ranking ok"


echo "start center,role..."
./centerserver &
./roleserver &

echo "start daer, room, mail, gmserver matchserver matchdaerserver..."
./daerserver &
./roomserver &
./mailserver &
./gmserver &
./matchserver &
./matchdaerserver &
./pockerserver &
./majiangserver&
./payserver&


echo "start gas..."
./gameserver -c ../cfg/gameserver1.json &
echo "start all ok!"

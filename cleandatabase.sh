mysqlu=root
mysqlp=ppedbs
#mysqlpath=
mysqlpath='/usr/local/mysql/bin/'

#mysql清除
${mysqlpath}mysql -u$mysqlu -p$mysqlp < server/sql/clean.sql
${mysqlpath}mysql -u$mysqlu -p$mysqlp < server/sql/game.sql

#redispath=
redispath=/usr/local/redis-3.2.5/src/
#redis清除
for i in $(seq 16);do
    ${redispath}redis-cli -p 6379 -n $(($i - 1)) flushdb
done

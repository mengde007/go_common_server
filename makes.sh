#!/bin/sh
#
#SVN路径 密码用户
SVN_PATH="svn://192.168.8.12:8001/Trunk/Data/"
SVN_USER="wuyuhuan"
SVN_PASS="xxxholic"
#工作目录

CUR_PATH="/root/shell/svn/Data"
[ -d $CUR_PATH ] || mkdir -p $CUR_PATH

Version_File="$CUR_PATH/x-Version.xml"
svn checkout --force --username=$SVN_USER --password=$SVN_PASS --non-interactive  $SVN_PATH $CUR_PATH

setverson() {
    svnv=`svnversion |sed 's/^.*://' |sed 's/[A-Z]*$//'`
    echo $svnv 

    regs="s/VersionNew\([[:punct:]]*\)[0-9]\{1,\}/VersionNew\1$svnv/"
    # client Config
    sed -i $regs $Version_File
    # gateServer Config
    sed -i $regs bin/cfg/gateserver.json
    # GameServer Config
    #sed -i $regs bin/gas1.json
    # sed -i $regs bin/gas2.json
    # sed -i $regs bin/gas3.json
    svn commit $Version_File -m "x-Version修改" --username=$SVN_USER --password=$SVN_PASS
    
    # make Server
    sh make.sh
}
resetverson() {
    svnv=`svnversion |sed 's/^.*://' |sed 's/[A-Z]*$//'`
    echo $svnv

    regs="s/VersionOld\([[:punct:]]*\)[0-9]\{1,\}/VersionOld\1$svnv/"
    # client Config
    sed -i $regs $Version_File
    # gateServer Config
    sed -i $regs bin/cfg/gateserver.json
    # GameServer Config
    #sed -i $regs bin/gas1.json
    #sed -i $regs bin/gas2.json
    #sed -i $regs bin/gas3.json
    
    # set VersionNew
    setverson

}

case "$1" in
    v)
        setverson
        ;;
    V)
        resetverson
        ;;
    *)
        echo $"Usage: $0 {v (unSetVerson)| V (reSetVerson)}"
        exit 2
esac

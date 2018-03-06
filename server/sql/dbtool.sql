drop database if exists db_ttdsg_playerdata;
create database db_ttdsg_playerdata;
use db_ttdsg_playerdata;

create table t_player_data
(
    OpenId                  		varchar(128)    not null,    #openid:平台（用":"间隔） 平台定义1：安卓，2：IOS
    Uid                  			varchar(128)    not null,    #游戏内id
    CharName     	    		    varchar(128)    not null,    #玩家名字
    OfficialLevel           		int unsigned    not null,    #太守府等级
    OfficialExp       	    		int unsigned    not null,    #战功
    Trophy		    			    int unsigned    not null,    #杯数
    CenterLevel             		int unsigned    not null,    #本数
    PveNormalStage	    		    int unsigned    not null,    #普通关卡
    PveHardStage	    		    int unsigned    not null,    #困难关卡
    PveNightmareStage	    	    int unsigned    not null,    #噩梦关卡

    Gold	            			int unsigned    not null,    #金
    Food		    				int unsigned    not null,    #粮
    ZiJin		    				int unsigned    not null,    #紫金
    Clan		    				varchar(128)    not null,    #联盟
    Heros                   		longtext        not null,    #英雄们  id,level多个武将之间用":"间隔
    TaskInfo                        longtext        not null,    #任务 id,finishedTime(没有完成时finishedTime为0), 多个之间用":"分割
    VillageId                       varchar(128)    not null,    #村庄ID 16进制
    
    primary key(OpenId)
)engine=innodb DEFAULT CHARSET utf8;

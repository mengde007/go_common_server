drop database if exists dpsglog;

create database dpsglog;

use	dpsglog;

#login log
create table tbl_playerlogin
(
	pl_AutoId      		bigint unsigned    not null   auto_increment, 
	pl_CharId			varchar(128)		default "", #玩家ID
	pl_DateTime			datetime			not null, 	#登陆时间
	pl_ChannelId		tinyint unsigned	not null default 0,#渠道号
	pl_Info				varchar(128)		default "", #登陆登出信息，登入: "0;IP 登出: 1;"

	primary key(pl_AutoId),
	key(pl_CharId)
)engine=innodb;

create table log_res_gain #资源获得
(
	pl_AutoId      bigint unsigned    not null   auto_increment, 
	pl_ChannelId		tinyint unsigned	not null default 0,#渠道号
	pl_UId         varchar(128)       not null   default "", #玩家ID
	pl_Time        datetime           not null,              #时间
	pl_ResType     varchar(128)       not null   default "", #资源类型
	pl_ResNum      int(10) unsigned   not null   default 0 , #数量
	pl_ResWay      int(10) unsigned   not null   default 0 , #来源途径
	
	primary key(pl_AutoId),
	key(pl_UId)
)engine=innodb;


create table log_res_lose #资源消耗
(
	pl_AutoId      bigint unsigned    not null   auto_increment, 
	pl_ChannelId		tinyint unsigned	not null default 0,#渠道号
	pl_UId         varchar(128)       not null   default "", #玩家ID
	pl_Time        datetime           not null,              #时间
	pl_ResType     varchar(128)       not null   default "", #资源类型
	pl_ResNum      int(10) unsigned   not null   default 0 , #数量
	pl_ResWay      int(10) unsigned   not null   default 0 , #花费途径
	
	primary key(pl_AutoId),
	key(pl_UId)
)engine=innodb;


create table log_taobao_pay #纪录淘宝消费纪录
(
	tp_UId         varchar(128)       not null   default "", #玩家ID
	pl_ChannelId   int unsigned	   not null   default 0,	#渠道号
	tp_TradeTime   datetime           not null,              #时间
	tp_TradeError  varchar(128)       not null   default "", #错误信息
	tp_TradeEnd    tinyint            not null, 				#交易结果
	tp_TradeNumber varchar(128)	   not null,
	tp_ItemName    varchar(128)	   not null,   			   #数量
	tp_TotoalPee   float   			   not null   default 0 , #价格
	
	key(tp_UId)
)engine=innodb;
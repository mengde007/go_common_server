CREATE DATABASE  IF NOT EXISTS `gmtool` ;
USE `gmtool`;


DROP TABLE IF EXISTS `dim_date`;

CREATE TABLE `dim_date` (
  `date_id` int(11) NOT NULL COMMENT '20110512',
  `date_name` varchar(16) DEFAULT NULL COMMENT '2011-05-12',
  `date_of_month` int(11) DEFAULT NULL COMMENT '12',
  `year_id` int(11) DEFAULT NULL COMMENT '2011',
  `year_name` varchar(16) DEFAULT NULL COMMENT '2011年',
  `quarter_id` int(11) DEFAULT NULL COMMENT '2',
  `quarter_name` varchar(16) DEFAULT NULL COMMENT '2季度',
  `month_id` int(11) DEFAULT NULL COMMENT '5',
  `month_name` varchar(16) DEFAULT NULL COMMENT '5月',
  `month_of_year_name` varchar(16) DEFAULT NULL COMMENT '2011年5月',
  `month_of_year_id` int(11) DEFAULT NULL COMMENT '201105',
  `week_id` int(11) DEFAULT NULL,
  `week_name` varchar(16) DEFAULT NULL,
  `week_of_year_id` int(11) DEFAULT NULL,
  `week_of_year_name` varchar(32) DEFAULT NULL,
  `is_weekend` enum('1','0') DEFAULT NULL COMMENT '是否周末',
  PRIMARY KEY (`date_id`),
  KEY `ix_dim_date_date_name` (`date_name`),
  KEY `ix_dim_date_month_id` (`month_id`),
  KEY `ix_dim_date_year_id` (`year_id`),
  KEY `ix_dim_date_quanter_id` (`quarter_id`),
  KEY `ix_dim_date_week_of_year_id` (`week_of_year_id`,`week_of_year_name`)
) ENGINE=MyISAM DEFAULT CHARSET=latin1;

LOCK TABLES `dim_date` WRITE;

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_churn_1day`;

CREATE TABLE `gm_churn_1day` (
  `day` char(8) NOT NULL COMMENT '日期',
  `channel` int(2) NOT NULL DEFAULT '0' COMMENT '渠道',
  `count` int(11) NOT NULL,
  `churn` int(11) NOT NULL,
  `ratio` float NOT NULL,
  PRIMARY KEY (`day`,`channel`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


LOCK TABLES `gm_churn_1day` WRITE;

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_churn_30day`;

CREATE TABLE `gm_churn_30day` (
  `day` char(8) NOT NULL COMMENT '日期',
  `channel` int(2) NOT NULL DEFAULT '0' COMMENT '渠道',
  `count` int(11) NOT NULL,
  `churn` int(11) NOT NULL,
  `ratio` float NOT NULL,
  PRIMARY KEY (`day`,`channel`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


LOCK TABLES `gm_churn_30day` WRITE;

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_churn_7day`;

CREATE TABLE `gm_churn_7day` (
  `day` char(8) NOT NULL COMMENT '日期',
  `channel` int(2) NOT NULL DEFAULT '0' COMMENT '渠道',
  `count` int(11) NOT NULL,
  `churn` int(11) NOT NULL,
  `ratio` float NOT NULL,
  PRIMARY KEY (`day`,`channel`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


LOCK TABLES `gm_churn_7day` WRITE;

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_email`;

CREATE TABLE `gm_email` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '邮件ID',
  `title` varchar(255) NOT NULL DEFAULT 'NoTitle' COMMENT '邮件title',
  `from` varchar(20) NOT NULL DEFAULT '0',
  `send_time` int(11) NOT NULL COMMENT '发送时间',
  `utc_time` char(25) NOT NULL,
  `duration` int(11) NOT NULL DEFAULT '86400',
  `status` int(11) NOT NULL DEFAULT '0',
  `content` varchar(255) NOT NULL,
  `mode` varchar(100) NOT NULL DEFAULT '0',
  `sum` int(11) NOT NULL DEFAULT '0',
  `channel` int(2) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='邮件表';


LOCK TABLES `gm_email` WRITE;

UNLOCK TABLES;


DROP TABLE IF EXISTS `gm_email_attachment`;

CREATE TABLE `gm_email_attachment` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `key` varchar(255) NOT NULL,
  `description` varchar(255) NOT NULL,
  `ctime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


LOCK TABLES `gm_email_attachment` WRITE;

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_email_data`;

CREATE TABLE `gm_email_data` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `mail_id` int(11) NOT NULL COMMENT '邮件id',
  `user_name` varchar(46) NOT NULL COMMENT '玩家id',
  `status` int(1) NOT NULL DEFAULT '0' COMMENT '发送状态',
  PRIMARY KEY (`id`),
  KEY `mail_id` (`mail_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='邮件发送状态';


LOCK TABLES `gm_email_data` WRITE;

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_email_mode`;

CREATE TABLE `gm_email_mode` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'mode-id',
  `value` varchar(22) NOT NULL COMMENT '对应游戏辨认KEY值',
  `name` varchar(25) NOT NULL COMMENT ' 中文名称',
  `description` varchar(255) NOT NULL COMMENT '描述',
  `ctime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00' ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=15 DEFAULT CHARSET=utf8;


LOCK TABLES `gm_email_mode` WRITE;

INSERT INTO `gm_email_mode` VALUES (1,'1','Food','粮草什么的','2013-12-27 07:15:06'),(2,'2','Gold','金币什么的','2013-12-27 07:18:08'),(3,'3','Gem','宝石','2014-01-06 02:14:56'),(4,'4','Wuhun','武魂什么的','2014-01-06 01:13:08'),(5,'51','Barbarian','','0000-00-00 00:00:00'),(6,'52','Archer','','0000-00-00 00:00:00'),(7,'53','Goblin','','0000-00-00 00:00:00'),(8,'54','Giant','','0000-00-00 00:00:00'),(9,'55','WallBreaker','','0000-00-00 00:00:00'),(10,'56','Balloon','','0000-00-00 00:00:00'),(11,'57','Wizard','','0000-00-00 00:00:00'),(12,'58','Healer','','0000-00-00 00:00:00'),(13,'59','Dragon','','0000-00-00 00:00:00'),(14,'60','PEKKA','','0000-00-00 00:00:00');

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_mode`;

CREATE TABLE `gm_mode` (
  `id` smallint(6) unsigned NOT NULL AUTO_INCREMENT,
  `name` char(40) NOT NULL DEFAULT '',
  `parentid` smallint(6) NOT NULL DEFAULT '0',
  `m` varchar(20) NOT NULL DEFAULT '',
  `a` varchar(20) NOT NULL DEFAULT '',
  `name2` varchar(40) NOT NULL DEFAULT '中文名',
  `listorder` smallint(6) unsigned NOT NULL DEFAULT '0',
  `display` enum('1','0') NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`),
  KEY `module` (`m`,`a`)
) ENGINE=MyISAM AUTO_INCREMENT=1709 DEFAULT CHARSET=utf8;


LOCK TABLES `gm_mode` WRITE;

INSERT INTO `gm_mode` VALUES (1669,'Dashboard',1,'admin','dashboard','欢迎界面',0,'1'),(1670,'Notice',1,'admin','notice','通知',2,'1'),(1671,'User',1,'admin','user','游戏用户管理',3,'1'),(1672,'Manage Email',1,'admin','mail','邮件管理',5,'1'),(1673,'Events Calendar',1,'admin','event','定时任务',8,'1'),(1674,'Settings',1,'admin','set','设置',7,'1'),(1675,'All Notice',1670,'notice','list','通知列表',0,'1'),(1676,'Add a new Notice',1670,'notice','add','添加通知',0,'1'),(1677,'Email List',1672,'mail','list','邮件列表',2,'1'),(1678,'Send a new Email',1672,'mail','add','发送邮件',1,'1'),(1679,'Email Attachment Settings',1672,'mail','set','设置邮件附件',0,'1'),(1680,'Users and Permissions',1674,'user','list','用户列表',0,'1'),(1681,'Add a new User',1671,'user','add','添加用户',0,'0'),(1682,'Edit User Info',1671,'user','edit','编辑用户',0,'0'),(1683,'Module Manage',1674,'mode','list','模块管理',1,'1'),(1689,'Role Manage',1674,'role','list','角色管理',2,'1'),(1693,'Find from DB',1671,'guser','serach','搜索用户',0,'1'),(1,'UI',0,'root','UI','界面',0,'0'),(1691,'Add Role',1674,'role','add','添加角色',2,'0'),(1694,'Edit Guser',1671,'guser','edit','编辑游戏用户',0,'0'),(1695,'Send To',1672,'mail','sendto','查看接收用户',3,'0'),(1696,'Sign Player',1669,'admin','splayer',' 登录用户',0,'1'),(1698,'Channel List',1674,'channel','list','渠道列表',3,'1'),(1700,'Change Site',1674,'channel','goto','切换通道 ',0,'0'),(1701,'Server Node',1,'admin','server','Server Node',6,'1'),(1702,'Event List',1701,'server','evtlist','Event List',3,'1'),(1703,'NodeType List',1701,'server','typelist','节点类型',4,'1'),(1704,'Node List',1701,'server','nodelist','服务节点',2,'1'),(1705,'Event Sets',1701,'server','evtsets','事件集',1,'1'),(1706,'BatchEvt',1673,'server','batchevt','事件批处理',0,'1'),(1708,'Change Config',1701,'server','changecfg','修改配置表',0,'1');

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_newuser`;

CREATE TABLE `gm_newuser` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `timeline` char(8) NOT NULL,
  `uid` char(36) NOT NULL,
  `channel` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `timeline` (`timeline`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


LOCK TABLES `gm_newuser` WRITE;

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_notice`;

CREATE TABLE `gm_notice` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '通知ID',
  `start_time` int(10) NOT NULL COMMENT '开始时间',
  `utc_time` char(25) NOT NULL,
  `time_line` int(10) NOT NULL COMMENT '时间间隔',
  `notice_sum` int(5) NOT NULL DEFAULT '1' COMMENT '通知次数',
  `type` int(1) NOT NULL DEFAULT '0' COMMENT '通知类型',
  `content` varchar(255) NOT NULL COMMENT '通知内容',
  `surplus` int(5) NOT NULL COMMENT '通知剩余量',
  `last_time` int(10) NOT NULL,
  `channel` int(2) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='通知';

LOCK TABLES `gm_notice` WRITE;

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_role`;

CREATE TABLE `gm_role` (
  `roleid` tinyint(3) unsigned NOT NULL AUTO_INCREMENT,
  `admin` int(1) NOT NULL DEFAULT '0',
  `rolename` varchar(50) NOT NULL,
  `description` text NOT NULL,
  `listorder` smallint(5) unsigned NOT NULL DEFAULT '0',
  `disabled` tinyint(1) unsigned NOT NULL DEFAULT '0',
  `siteid` int(11) NOT NULL,
  PRIMARY KEY (`roleid`),
  KEY `listorder` (`listorder`),
  KEY `disabled` (`disabled`)
) ENGINE=MyISAM AUTO_INCREMENT=14 DEFAULT CHARSET=utf8;


LOCK TABLES `gm_role` WRITE;

INSERT INTO `gm_role` VALUES (1,1,'超级管理员','超级管理员',0,0,1),(2,0,'站点管理员','站点管理员',0,0,1),(3,0,'运营总监','运营总监',1,0,1),(4,0,'总编','总编',5,0,1),(10,1,'超级管理员','超级管理员',0,0,2),(11,1,'root','root',0,1,6),(13,1,'admin','root',0,1,5);

UNLOCK TABLES;


DROP TABLE IF EXISTS `gm_role_priv`;

CREATE TABLE `gm_role_priv` (
  `catid` smallint(5) unsigned NOT NULL DEFAULT '0',
  `roleid` smallint(5) unsigned NOT NULL DEFAULT '0',
  `m` varchar(45) NOT NULL,
  `a` varchar(45) NOT NULL,
  `siteid` int(11) NOT NULL DEFAULT '1'
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


LOCK TABLES `gm_role_priv` WRITE;

INSERT INTO `gm_role_priv` VALUES (2,2,'','event',2),(0,2,'a9c2cd5f00b2cd28cbb1eb2eaeb412ba8da1fef5','',1),(0,2,'4c547c574d1846e1fe5efc345db66fbce63bcf94','',1),(0,2,'771b31b0da1cf5916fd802c895309169ac13af5c','',1),(0,2,'39317d0290dd234b5f45db6cfb80c56d11014af0','',1),(0,2,'75dcf9f9bc71c10f67377f2b52febd3720ad27f8','',1),(0,2,'c62322644f3c16836aabc38b8071fc33ce384038','',1),(0,2,'21e9fe2788eb5330e7882c5f37df8035b7879d7d','',1),(0,2,'1692d052bd216f607432f928c5f5c41597ed1f03','',1),(0,2,'d4746442e291e44d3346eaa4293b9cd6c9745903','',1),(0,2,'67bd0627285de34e123f8bafea30b3215216a87b','',1),(0,2,'67bc0bf05304388d85436d597561f08519ba7401','',1),(0,2,'35e7cf5dadbda483383fc531ee0438ee766fff34','',1),(0,2,'f00b51fbde6bbc7b273639796f0e9055e7c41086','',1),(0,2,'5d617fb9171eb26e965333ce1a2a3326c74e18ed','',1),(0,2,'6e18eeeeae7c0c5a9d0b4ae8c8b95f6609793239','',1),(0,2,'017aad1159ef65fd995b95ccd427b718a0787635','',1),(0,2,'8409a60f70eb0cfcc4745033d8d83a267c6c5373','',1),(0,2,'20f7a371795d875e5509b08bee0959ec98382756','',1),(0,2,'043382b43bd056954094a94f05c864bb5ca845e0','',1),(0,2,'ebea76299b9ea1dbcb218b7b1b2e104d7a6813fb','',1),(0,10,'71fa8b1980bd889b4f8eff91b3a55cd78a8d80e3','',2),(0,10,'a9c2cd5f00b2cd28cbb1eb2eaeb412ba8da1fef5','',2),(0,10,'4c547c574d1846e1fe5efc345db66fbce63bcf94','',2),(0,10,'771b31b0da1cf5916fd802c895309169ac13af5c','',2),(0,10,'75dcf9f9bc71c10f67377f2b52febd3720ad27f8','',2),(0,10,'c62322644f3c16836aabc38b8071fc33ce384038','',2),(0,10,'21e9fe2788eb5330e7882c5f37df8035b7879d7d','',2),(0,10,'1692d052bd216f607432f928c5f5c41597ed1f03','',2),(0,10,'d4746442e291e44d3346eaa4293b9cd6c9745903','',2),(0,10,'67bd0627285de34e123f8bafea30b3215216a87b','',2),(0,10,'67bc0bf05304388d85436d597561f08519ba7401','',2),(0,10,'35e7cf5dadbda483383fc531ee0438ee766fff34','',2),(0,10,'f00b51fbde6bbc7b273639796f0e9055e7c41086','',2),(0,10,'6e18eeeeae7c0c5a9d0b4ae8c8b95f6609793239','',2),(0,10,'017aad1159ef65fd995b95ccd427b718a0787635','',2),(0,10,'8409a60f70eb0cfcc4745033d8d83a267c6c5373','',2),(0,10,'20f7a371795d875e5509b08bee0959ec98382756','',2),(0,10,'043382b43bd056954094a94f05c864bb5ca845e0','',2),(0,10,'bf6450ef57371ae2cd7bfc8b46964b359f2ced4f','',2),(0,11,'','',4),(0,11,'71fa8b1980bd889b4f8eff91b3a55cd78a8d80e3','',4),(0,11,'a9c2cd5f00b2cd28cbb1eb2eaeb412ba8da1fef5','',4),(0,11,'4c547c574d1846e1fe5efc345db66fbce63bcf94','',4),(0,11,'771b31b0da1cf5916fd802c895309169ac13af5c','',4),(0,11,'75dcf9f9bc71c10f67377f2b52febd3720ad27f8','',4),(0,11,'c62322644f3c16836aabc38b8071fc33ce384038','',4),(0,11,'21e9fe2788eb5330e7882c5f37df8035b7879d7d','',4),(0,11,'1692d052bd216f607432f928c5f5c41597ed1f03','',4),(0,11,'d4746442e291e44d3346eaa4293b9cd6c9745903','',4),(0,11,'67bd0627285de34e123f8bafea30b3215216a87b','',4),(0,11,'67bc0bf05304388d85436d597561f08519ba7401','',4),(0,11,'35e7cf5dadbda483383fc531ee0438ee766fff34','',4),(0,11,'f00b51fbde6bbc7b273639796f0e9055e7c41086','',4),(0,14,'6e18eeeeae7c0c5a9d0b4ae8c8b95f6609793239','',4),(0,11,'017aad1159ef65fd995b95ccd427b718a0787635','',4),(0,11,'8409a60f70eb0cfcc4745033d8d83a267c6c5373','',4),(0,11,'20f7a371795d875e5509b08bee0959ec98382756','',4),(0,11,'043382b43bd056954094a94f05c864bb5ca845e0','',4),(0,11,'bf6450ef57371ae2cd7bfc8b46964b359f2ced4f','',4),(0,0,'71fa8b1980bd889b4f8eff91b3a55cd78a8d80e3','',6),(0,0,'a9c2cd5f00b2cd28cbb1eb2eaeb412ba8da1fef5','',6),(0,0,'4c547c574d1846e1fe5efc345db66fbce63bcf94','',6),(0,0,'771b31b0da1cf5916fd802c895309169ac13af5c','',6),(0,0,'75dcf9f9bc71c10f67377f2b52febd3720ad27f8','',6),(0,0,'67bc0bf05304388d85436d597561f08519ba7401','',6),(0,0,'5d617fb9171eb26e965333ce1a2a3326c74e18ed','',6),(0,0,'6e18eeeeae7c0c5a9d0b4ae8c8b95f6609793239','',6),(0,0,'8409a60f70eb0cfcc4745033d8d83a267c6c5373','',6),(0,0,'40ca03fdbf1d88a9d74635d67caff90c800d801d','',6),(0,0,'71fa8b1980bd889b4f8eff91b3a55cd78a8d80e3','',5),(0,0,'a9c2cd5f00b2cd28cbb1eb2eaeb412ba8da1fef5','',5),(0,0,'4c547c574d1846e1fe5efc345db66fbce63bcf94','',5),(0,0,'771b31b0da1cf5916fd802c895309169ac13af5c','',5),(0,0,'75dcf9f9bc71c10f67377f2b52febd3720ad27f8','',5),(0,0,'c62322644f3c16836aabc38b8071fc33ce384038','',5),(0,0,'21e9fe2788eb5330e7882c5f37df8035b7879d7d','',5),(0,0,'1692d052bd216f607432f928c5f5c41597ed1f03','',5),(0,0,'d4746442e291e44d3346eaa4293b9cd6c9745903','',5),(0,0,'67bc0bf05304388d85436d597561f08519ba7401','',5),(0,0,'35e7cf5dadbda483383fc531ee0438ee766fff34','',5),(0,0,'f00b51fbde6bbc7b273639796f0e9055e7c41086','',5),(0,0,'6e18eeeeae7c0c5a9d0b4ae8c8b95f6609793239','',5),(0,0,'017aad1159ef65fd995b95ccd427b718a0787635','',5),(0,0,'8409a60f70eb0cfcc4745033d8d83a267c6c5373','',5),(0,0,'20f7a371795d875e5509b08bee0959ec98382756','',5),(0,0,'043382b43bd056954094a94f05c864bb5ca845e0','',5),(0,0,'0f25210b3a8cfa4ce3f318ba9274bb6d7104848c','',5),(0,13,'71fa8b1980bd889b4f8eff91b3a55cd78a8d80e3','',5),(0,13,'a9c2cd5f00b2cd28cbb1eb2eaeb412ba8da1fef5','',5),(0,13,'4c547c574d1846e1fe5efc345db66fbce63bcf94','',5),(0,13,'771b31b0da1cf5916fd802c895309169ac13af5c','',5),(0,13,'75dcf9f9bc71c10f67377f2b52febd3720ad27f8','',5),(0,13,'c62322644f3c16836aabc38b8071fc33ce384038','',5),(0,13,'21e9fe2788eb5330e7882c5f37df8035b7879d7d','',5),(0,13,'1692d052bd216f607432f928c5f5c41597ed1f03','',5),(0,13,'d4746442e291e44d3346eaa4293b9cd6c9745903','',5),(0,13,'67bc0bf05304388d85436d597561f08519ba7401','',5),(0,13,'35e7cf5dadbda483383fc531ee0438ee766fff34','',5),(0,13,'f00b51fbde6bbc7b273639796f0e9055e7c41086','',5),(0,13,'6e18eeeeae7c0c5a9d0b4ae8c8b95f6609793239','',5),(0,13,'017aad1159ef65fd995b95ccd427b718a0787635','',5),(0,13,'8409a60f70eb0cfcc4745033d8d83a267c6c5373','',5),(0,13,'20f7a371795d875e5509b08bee0959ec98382756','',5),(0,13,'043382b43bd056954094a94f05c864bb5ca845e0','',5),(0,13,'0f25210b3a8cfa4ce3f318ba9274bb6d7104848c','',5),(0,11,'71fa8b1980bd889b4f8eff91b3a55cd78a8d80e3','',6),(0,11,'4c547c574d1846e1fe5efc345db66fbce63bcf94','',6),(0,11,'771b31b0da1cf5916fd802c895309169ac13af5c','',6),(0,11,'1692d052bd216f607432f928c5f5c41597ed1f03','',6),(0,11,'d4746442e291e44d3346eaa4293b9cd6c9745903','',6),(0,11,'67bd0627285de34e123f8bafea30b3215216a87b','',6),(0,11,'35e7cf5dadbda483383fc531ee0438ee766fff34','',6),(0,11,'f00b51fbde6bbc7b273639796f0e9055e7c41086','',6),(0,11,'017aad1159ef65fd995b95ccd427b718a0787635','',6),(0,11,'20f7a371795d875e5509b08bee0959ec98382756','',6),(0,11,'043382b43bd056954094a94f05c864bb5ca845e0','',6),(0,4,'a9c2cd5f00b2cd28cbb1eb2eaeb412ba8da1fef5','',1),(0,4,'4c547c574d1846e1fe5efc345db66fbce63bcf94','',1),(0,1,'71fa8b1980bd889b4f8eff91b3a55cd78a8d80e3','',1),(0,1,'a9c2cd5f00b2cd28cbb1eb2eaeb412ba8da1fef5','',1),(0,1,'4c547c574d1846e1fe5efc345db66fbce63bcf94','',1),(0,1,'771b31b0da1cf5916fd802c895309169ac13af5c','',1),(0,1,'39317d0290dd234b5f45db6cfb80c56d11014af0','',1),(0,1,'75dcf9f9bc71c10f67377f2b52febd3720ad27f8','',1),(0,1,'c62322644f3c16836aabc38b8071fc33ce384038','',1),(0,1,'21e9fe2788eb5330e7882c5f37df8035b7879d7d','',1),(0,1,'1692d052bd216f607432f928c5f5c41597ed1f03','',1),(0,1,'d4746442e291e44d3346eaa4293b9cd6c9745903','',1),(0,1,'67bd0627285de34e123f8bafea30b3215216a87b','',1),(0,1,'67bc0bf05304388d85436d597561f08519ba7401','',1),(0,1,'35e7cf5dadbda483383fc531ee0438ee766fff34','',1),(0,1,'f00b51fbde6bbc7b273639796f0e9055e7c41086','',1),(0,1,'5d617fb9171eb26e965333ce1a2a3326c74e18ed','',1),(0,1,'6e18eeeeae7c0c5a9d0b4ae8c8b95f6609793239','',1),(0,1,'017aad1159ef65fd995b95ccd427b718a0787635','',1),(0,1,'8409a60f70eb0cfcc4745033d8d83a267c6c5373','',1),(0,1,'20f7a371795d875e5509b08bee0959ec98382756','',1),(0,1,'043382b43bd056954094a94f05c864bb5ca845e0','',1),(0,1,'0f25210b3a8cfa4ce3f318ba9274bb6d7104848c','',1),(0,1,'ebea76299b9ea1dbcb218b7b1b2e104d7a6813fb','',1),(0,1,'40ca03fdbf1d88a9d74635d67caff90c800d801d','',1),(0,1,'f636c8e5424e1350ed292c2dac865ae4ab3c7c49','',1),(0,1,'936479caff0dde8586c5b09166e221eee054bb53','',1),(0,1,'fe795965a9dd4a89350964d1e9bb8217afec9eba','',1),(0,1,'0abdd22f591021f1e4eb8ca9b18257914e1bbbf5','',1),(0,1,'44fa8142126915b48bb5be9d7488434a73785da3','',1),(0,1,'9b3ebec13ad5a2ec75c6c1084058b8cc32b834ed','',1),(0,1,'23f3b069ee97cef273367999ba11021bd5e8377d','',1);

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_role_priv3`;

CREATE TABLE `gm_role_priv3` (
  `catid` smallint(5) unsigned NOT NULL DEFAULT '0',
  `roleid` smallint(5) unsigned NOT NULL DEFAULT '0',
  `action` char(30) NOT NULL,
  `siteid` int(11) NOT NULL DEFAULT '0',
  KEY `catid` (`catid`,`roleid`,`action`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


LOCK TABLES `gm_role_priv3` WRITE;

INSERT INTO `gm_role_priv3` VALUES (8,1,'remove',1),(8,1,'push',1),(8,1,'listorder',1),(8,1,'delete',1),(8,1,'edit',1),(8,1,'add',1),(8,1,'init',1),(6,1,'remove',1),(6,1,'push',1),(6,1,'listorder',1),(6,1,'delete',1),(6,1,'edit',1),(6,4,'add',1),(6,1,'init',1),(4,1,'init',1),(3,1,'init',1),(1,1,'init',1),(13,1,'remove',1),(13,1,'push',1),(13,1,'listorder',1),(13,1,'delete',1),(13,1,'edit',1),(13,1,'add',1),(13,1,'init',1),(10,1,'init',1),(11,1,'init',1),(8,1,'remove',1),(8,1,'push',1),(8,1,'listorder',1),(8,1,'delete',1),(8,1,'edit',1),(8,1,'add',1),(8,1,'init',1),(6,1,'remove',1),(6,1,'push',1),(6,1,'listorder',1),(6,1,'delete',1),(6,1,'edit',1),(1681,1,'add',1),(6,1,'init',1),(4,1,'init',1),(3,1,'init',1),(1,1,'init',1),(13,1,'remove',1),(13,1,'push',1),(13,1,'listorder',1),(13,1,'delete',1),(13,1,'edit',1),(13,1,'add',1),(13,1,'init',1),(10,1,'init',1),(11,1,'init',1),(3,2,'xvzx',1),(3,2,'xvzx',1),(1670,2,'notice',1),(1671,1,'user',1),(1671,1,'user',1),(1669,4,'dashboard',1),(1670,2,'notice',1),(1671,2,'user',1),(1670,2,'notice',1),(1670,2,'notice',1),(1670,2,'notice',1),(1671,2,'user',1),(1672,2,'mail',1),(1669,1,'dashboard',1);

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_sign_1day`;

CREATE TABLE `gm_sign_1day` (
  `day` int(8) NOT NULL COMMENT '日期',
  `channel` int(2) NOT NULL DEFAULT '0' COMMENT '渠道',
  `count` int(11) NOT NULL COMMENT '一天前新增人数',
  `retention` int(11) NOT NULL COMMENT '今日登录人数',
  `ratio` float NOT NULL DEFAULT '0',
  PRIMARY KEY (`day`,`channel`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


LOCK TABLES `gm_sign_1day` WRITE;

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_sign_30day`;

CREATE TABLE `gm_sign_30day` (
  `day` int(8) NOT NULL COMMENT '日期',
  `channel` int(2) NOT NULL DEFAULT '0' COMMENT '渠道',
  `count` int(11) NOT NULL COMMENT '一天前新增人数',
  `retention` int(11) NOT NULL COMMENT '今日登录人数',
  `ratio` float NOT NULL DEFAULT '0',
  PRIMARY KEY (`day`,`channel`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


LOCK TABLES `gm_sign_30day` WRITE;

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_sign_7day`;

CREATE TABLE `gm_sign_7day` (
  `day` int(8) NOT NULL COMMENT '日期',
  `channel` int(2) NOT NULL DEFAULT '0' COMMENT '渠道',
  `count` int(11) NOT NULL COMMENT '一天前新增人数',
  `retention` int(11) NOT NULL COMMENT '今日登录人数',
  `ratio` float NOT NULL DEFAULT '0',
  PRIMARY KEY (`day`,`channel`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


LOCK TABLES `gm_sign_7day` WRITE;

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_sign_cache`;

CREATE TABLE `gm_sign_cache` (
  `uid` char(36) NOT NULL COMMENT '用户ID ',
  `hh` char(2) NOT NULL COMMENT '登录时间',
  `channel` int(2) NOT NULL DEFAULT '0' COMMENT '渠道号',
  `day` char(10) NOT NULL COMMENT ' 时间',
  PRIMARY KEY (`uid`,`hh`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


LOCK TABLES `gm_sign_cache` WRITE;

UNLOCK TABLES;

--
-- Table structure for table `gm_sign_timeline`
--

DROP TABLE IF EXISTS `gm_sign_timeline`;

CREATE TABLE `gm_sign_timeline` (
  `dayhh` char(10) NOT NULL,
  `day` char(10) NOT NULL COMMENT ' 时间',
  `hh` char(2) NOT NULL COMMENT '登录时间',
  `uid` char(36) NOT NULL COMMENT '用户ID ',
  `channel` int(2) NOT NULL DEFAULT '0' COMMENT '渠道号',
  PRIMARY KEY (`day`,`hh`,`uid`),
  KEY `dayhh` (`dayhh`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


LOCK TABLES `gm_sign_timeline` WRITE;

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_site`;

CREATE TABLE `gm_site` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(25) NOT NULL,
  `ename` varchar(40) DEFAULT NULL,
  `channel` int(11) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8;

LOCK TABLES `gm_site` WRITE;

INSERT INTO `gm_site` VALUES (1,'内部','root',0),(2,'test','test',2),(3,'越南','越南  ',9),(4,'越南','越南  ',9),(5,'QQ','QQ ',7),(6,'越南','越南  ',9);

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_site_email_attachment`;

CREATE TABLE `gm_site_email_attachment` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `key` int(11) NOT NULL,
  `num` int(11) NOT NULL DEFAULT '0',
  `channel` int(11) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=37 DEFAULT CHARSET=utf8;


LOCK TABLES `gm_site_email_attachment` WRITE;

INSERT INTO `gm_site_email_attachment` VALUES (3,1,0,1),(4,1,0,1),(24,2,65536,6),(25,3,65536,6),(26,5,65536,6),(27,6,65536,6),(28,2,65536,0),(29,3,65536,0),(30,5,65536,0),(31,6,65536,0),(32,2,65536,7),(33,6,65536,7),(34,3,65536,3),(35,6,65536,3),(36,2,65536,2);

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_site_mode`;

CREATE TABLE `gm_site_mode` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `mid` int(11) NOT NULL,
  `siteid` int(2) NOT NULL COMMENT '渠道ID',
  `m` varchar(45) NOT NULL,
  `a` varchar(45) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=MyISAM AUTO_INCREMENT=477 DEFAULT CHARSET=utf8;


LOCK TABLES `gm_site_mode` WRITE;

INSERT INTO `gm_site_mode` VALUES (476,1708,1,'23f3b069ee97cef273367999ba11021bd5e8377d',''),(475,1706,1,'9b3ebec13ad5a2ec75c6c1084058b8cc32b834ed',''),(474,1705,1,'44fa8142126915b48bb5be9d7488434a73785da3',''),(473,1704,1,'0abdd22f591021f1e4eb8ca9b18257914e1bbbf5',''),(254,1682,6,'f00b51fbde6bbc7b273639796f0e9055e7c41086',''),(253,1681,6,'35e7cf5dadbda483383fc531ee0438ee766fff34',''),(121,1696,2,'bf6450ef57371ae2cd7bfc8b46964b359f2ced4f',''),(120,1695,2,'043382b43bd056954094a94f05c864bb5ca845e0',''),(119,1694,2,'20f7a371795d875e5509b08bee0959ec98382756',''),(118,1691,2,'8409a60f70eb0cfcc4745033d8d83a267c6c5373',''),(117,1693,2,'017aad1159ef65fd995b95ccd427b718a0787635',''),(116,1689,2,'6e18eeeeae7c0c5a9d0b4ae8c8b95f6609793239',''),(115,1682,2,'f00b51fbde6bbc7b273639796f0e9055e7c41086',''),(114,1681,2,'35e7cf5dadbda483383fc531ee0438ee766fff34',''),(113,1680,2,'67bc0bf05304388d85436d597561f08519ba7401',''),(112,1679,2,'67bd0627285de34e123f8bafea30b3215216a87b',''),(111,1678,2,'d4746442e291e44d3346eaa4293b9cd6c9745903',''),(110,1677,2,'1692d052bd216f607432f928c5f5c41597ed1f03',''),(109,1676,2,'21e9fe2788eb5330e7882c5f37df8035b7879d7d',''),(108,1675,2,'c62322644f3c16836aabc38b8071fc33ce384038',''),(107,1674,2,'75dcf9f9bc71c10f67377f2b52febd3720ad27f8',''),(106,1672,2,'771b31b0da1cf5916fd802c895309169ac13af5c',''),(105,1671,2,'4c547c574d1846e1fe5efc345db66fbce63bcf94',''),(104,1670,2,'a9c2cd5f00b2cd28cbb1eb2eaeb412ba8da1fef5',''),(103,1669,2,'71fa8b1980bd889b4f8eff91b3a55cd78a8d80e3',''),(472,1703,1,'fe795965a9dd4a89350964d1e9bb8217afec9eba',''),(471,1702,1,'936479caff0dde8586c5b09166e221eee054bb53',''),(470,1701,1,'f636c8e5424e1350ed292c2dac865ae4ab3c7c49',''),(469,1700,1,'40ca03fdbf1d88a9d74635d67caff90c800d801d',''),(468,1698,1,'ebea76299b9ea1dbcb218b7b1b2e104d7a6813fb',''),(467,1696,1,'0f25210b3a8cfa4ce3f318ba9274bb6d7104848c',''),(466,1695,1,'043382b43bd056954094a94f05c864bb5ca845e0',''),(465,1694,1,'20f7a371795d875e5509b08bee0959ec98382756',''),(464,1691,1,'8409a60f70eb0cfcc4745033d8d83a267c6c5373',''),(463,1693,1,'017aad1159ef65fd995b95ccd427b718a0787635',''),(462,1689,1,'6e18eeeeae7c0c5a9d0b4ae8c8b95f6609793239',''),(461,1683,1,'5d617fb9171eb26e965333ce1a2a3326c74e18ed',''),(460,1682,1,'f00b51fbde6bbc7b273639796f0e9055e7c41086',''),(459,1681,1,'35e7cf5dadbda483383fc531ee0438ee766fff34',''),(458,1680,1,'67bc0bf05304388d85436d597561f08519ba7401',''),(457,1679,1,'67bd0627285de34e123f8bafea30b3215216a87b',''),(456,1678,1,'d4746442e291e44d3346eaa4293b9cd6c9745903',''),(455,1677,1,'1692d052bd216f607432f928c5f5c41597ed1f03',''),(454,1676,1,'21e9fe2788eb5330e7882c5f37df8035b7879d7d',''),(252,1679,6,'67bd0627285de34e123f8bafea30b3215216a87b',''),(251,1678,6,'d4746442e291e44d3346eaa4293b9cd6c9745903',''),(250,1677,6,'1692d052bd216f607432f928c5f5c41597ed1f03',''),(249,1672,6,'771b31b0da1cf5916fd802c895309169ac13af5c',''),(248,1671,6,'4c547c574d1846e1fe5efc345db66fbce63bcf94',''),(247,1669,6,'71fa8b1980bd889b4f8eff91b3a55cd78a8d80e3',''),(215,1669,5,'71fa8b1980bd889b4f8eff91b3a55cd78a8d80e3',''),(216,1670,5,'a9c2cd5f00b2cd28cbb1eb2eaeb412ba8da1fef5',''),(217,1671,5,'4c547c574d1846e1fe5efc345db66fbce63bcf94',''),(218,1672,5,'771b31b0da1cf5916fd802c895309169ac13af5c',''),(219,1674,5,'75dcf9f9bc71c10f67377f2b52febd3720ad27f8',''),(220,1675,5,'c62322644f3c16836aabc38b8071fc33ce384038',''),(221,1676,5,'21e9fe2788eb5330e7882c5f37df8035b7879d7d',''),(222,1677,5,'1692d052bd216f607432f928c5f5c41597ed1f03',''),(223,1678,5,'d4746442e291e44d3346eaa4293b9cd6c9745903',''),(224,1680,5,'67bc0bf05304388d85436d597561f08519ba7401',''),(225,1681,5,'35e7cf5dadbda483383fc531ee0438ee766fff34',''),(226,1682,5,'f00b51fbde6bbc7b273639796f0e9055e7c41086',''),(227,1689,5,'6e18eeeeae7c0c5a9d0b4ae8c8b95f6609793239',''),(228,1693,5,'017aad1159ef65fd995b95ccd427b718a0787635',''),(229,1691,5,'8409a60f70eb0cfcc4745033d8d83a267c6c5373',''),(230,1694,5,'20f7a371795d875e5509b08bee0959ec98382756',''),(231,1695,5,'043382b43bd056954094a94f05c864bb5ca845e0',''),(232,1696,5,'0f25210b3a8cfa4ce3f318ba9274bb6d7104848c',''),(255,1693,6,'017aad1159ef65fd995b95ccd427b718a0787635',''),(256,1694,6,'20f7a371795d875e5509b08bee0959ec98382756',''),(257,1695,6,'043382b43bd056954094a94f05c864bb5ca845e0',''),(453,1675,1,'c62322644f3c16836aabc38b8071fc33ce384038',''),(452,1674,1,'75dcf9f9bc71c10f67377f2b52febd3720ad27f8',''),(451,1673,1,'39317d0290dd234b5f45db6cfb80c56d11014af0',''),(450,1672,1,'771b31b0da1cf5916fd802c895309169ac13af5c',''),(449,1671,1,'4c547c574d1846e1fe5efc345db66fbce63bcf94',''),(448,1670,1,'a9c2cd5f00b2cd28cbb1eb2eaeb412ba8da1fef5',''),(447,1669,1,'71fa8b1980bd889b4f8eff91b3a55cd78a8d80e3','');

UNLOCK TABLES;



DROP TABLE IF EXISTS `gm_time_cache`;

CREATE TABLE `gm_time_cache` (
  `key` varchar(80) NOT NULL COMMENT '缓存数据名',
  `body` varchar(255) NOT NULL COMMENT '缓存内容',
  UNIQUE KEY `key_2` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;





DROP TABLE IF EXISTS `guser`;

CREATE TABLE `guser` (
 
  `uid` varchar(128) NOT NULL,
  `vid` varchar(28) NOT NULL DEFAULT '0' COMMENT 'villageId',
  `name` varchar(255) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '暂无' COMMENT '名称',
  `level` int(11) NOT NULL DEFAULT '0' COMMENT '等级',
  `clan` varchar(255) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '暂无' COMMENT '联盟',
  `diamonds` int(11) NOT NULL DEFAULT '0' COMMENT '宝石',
  `lose_diamonds` int(11) NOT NULL DEFAULT '0',
  `food` int(11) NOT NULL DEFAULT '0' COMMENT '粮草',
  `gold` int(11) NOT NULL DEFAULT '0' COMMENT '银币？',
  `wuhun` int(11) NOT NULL DEFAULT '0' COMMENT '武魂',
  `trophy` int(11) NOT NULL DEFAULT '0' COMMENT '令牌数',
  `drill_times` int(11) NOT NULL DEFAULT '0' COMMENT '演习次数',
  `center_level` int(11) NOT NULL DEFAULT '0' COMMENT '主营等级',
  `last_login` varchar(30) NOT NULL DEFAULT '0',
  `login_stats` varchar(50) NOT NULL DEFAULT '0' COMMENT '登录状态',
  `updated` varchar(20) NOT NULL DEFAULT '0',
  `channel` int(11) NOT NULL DEFAULT '0',
  `crtime` timestamp NOT NULL DEFAULT now(),
  PRIMARY KEY (`uid`),
  KEY `vid` (`vid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


LOCK TABLES `guser` WRITE;

UNLOCK TABLES;



DROP TABLE IF EXISTS `link`;

CREATE TABLE `link` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` char(50) NOT NULL,
  `url` char(255) NOT NULL,
  `description` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


LOCK TABLES `link` WRITE;

UNLOCK TABLES;



DROP TABLE IF EXISTS `sv_evtsets`;

CREATE TABLE `sv_evtsets` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(45) DEFAULT NULL,
  `body` varchar(255) DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8;


LOCK TABLES `sv_evtsets` WRITE;

INSERT INTO `sv_evtsets` VALUES (2,'Start All','MTozLDI6NA==','开启所有服务'),(3,'test2','Mjo0','test22'),(4,'mysqlNode全面检查','MTo2LDE6MTEsMToxNCwxOjEzLDE6MTIsMToxNSwxOjE3LDE6NywxOjgsMTo5LDE6MTAsMToxNg==','All event'),(5,'Top','MjoxOA==','Top');

UNLOCK TABLES;



DROP TABLE IF EXISTS `sv_nodeeventcache`;

CREATE TABLE `sv_nodeeventcache` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `nodeid` int(11) NOT NULL,
  `evtid` int(11) NOT NULL,
  `body` blob NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


LOCK TABLES `sv_nodeeventcache` WRITE;

UNLOCK TABLES;



DROP TABLE IF EXISTS `sv_nodelist`;

CREATE TABLE `sv_nodelist` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(45) NOT NULL,
  `host` varchar(15) NOT NULL,
  `user` varchar(45) NOT NULL,
  `typeid` int(11) NOT NULL,
  `pass` varchar(45) DEFAULT NULL,
  `port` int(5) DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL,
  `config` varchar(255) DEFAULT NULL,
  `log` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;


LOCK TABLES `sv_nodelist` WRITE;

INSERT INTO `sv_nodelist` VALUES (1,'mysql1','192.168.8.27','root',1,'dpsgdev!23',8889,'mysql  服务器                      ','var/bin',NULL),(2,'redis1','192.168.8.1','root',2,'dss',34,'redis 服务  ',NULL,NULL);

UNLOCK TABLES;



DROP TABLE IF EXISTS `sv_nodetype`;

CREATE TABLE `sv_nodetype` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(45) NOT NULL,
  `body` varchar(255) DEFAULT NULL,
  `Description` varchar(255) DEFAULT NULL,
  `dir` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;


LOCK TABLES `sv_nodetype` WRITE;

INSERT INTO `sv_nodetype` VALUES (1,'Mysql Server ','Myw1LDYsNyw4LDksMTAsMTEsMTIsMTMsMTQsMTUsMTYsMTc=','Mysql Server           ','fgbdfg'),(2,'Redis','NCwxOA==','Redis Server    ','var/www');

UNLOCK TABLES;



DROP TABLE IF EXISTS `sv_nodevent`;

CREATE TABLE `sv_nodevent` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(45) NOT NULL,
  `body` varchar(255) NOT NULL,
  `isloops` int(1) DEFAULT '0',
  `Duration` int(11) DEFAULT '0',
  `Description` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=19 DEFAULT CHARSET=utf8;


LOCK TABLES `sv_nodevent` WRITE;

INSERT INTO `sv_nodevent` VALUES (3,'start App','IC9iaW4vc2ggLWMgJ2NkIC5AVHlwZURpckA7IC4vQEFwcE5hbWVAID4vZGV2L251bGwgJiBzbGVlcCAxNXMgOyBwcyBhdSB8IGdyZXAgQEFwcE5hbWVAIHwgZ3JlcCAtdiBncmVwIDsgZWNobyAkPyA7IHRhaWwgLTUgQExvZ0An',0,0,'start all App Function    '),(4,'start Redis','c2FkZmFzZGZhcw==',0,0,'start Redis  '),(5,'stop Mysql','Y2VzZGE=',0,0,'STOP MYSQL'),(6,'检查系统版本','L2Jpbi9zaCAtYyAnZWNobyAwOyB1bmFtZSAtYTsn',0,0,'检查系统版本  '),(7,'磁盘硬件信息','L2Jpbi9zaCAtYyAnZWNobyAwOyBoZHBhcm0gLUkgL2Rldi9zZGE7Jw==',0,0,'磁盘硬件信息'),(8,'磁盘使用状况','L2Jpbi9zaCAtYyAnZWNobyAwOyBkZiAtaDsn',0,0,'磁盘配置分区使用状况'),(9,'查看网卡配置','L2Jpbi9zaCAtYyAnZWNobyAwOyBsc3BjaSB8IGdyZXAgLWkgZXRoOyc=',0,0,'查看网卡配置'),(10,'网络路由等配置','L2Jpbi9zaCAtYyAnZWNobyAwOyBuZXRzdGF0IC1ucjsn',0,0,'网络路由等配置'),(11,'CPU硬件信息','L2Jpbi9zaCAtYyAnZWNobyAwOyBkbWlkZWNvZGUgLXQgcHJvY2Vzc29yIDsn',0,0,'查看CPU硬件配置信息'),(12,'查看内存配置','L2Jpbi9zaCAtYyAnZWNobyAwOyBjYXQgL3Byb2MvbWVtaW5mbyA7Jw==',0,0,'查看内存配置'),(13,'查看内存使用','L2Jpbi9zaCAtYyAnZWNobyAwOyBmcmVlIC1tOyc=',0,0,'查看内存使用'),(14,'最后登录用户','L2Jpbi9zaCAtYyAnZWNobyAwOyBsYXN0IC1uIDIwOyc=',0,0,'最后登录用户'),(15,'最后重启时间','L2Jpbi9zaCAtYyAnZWNobyAwOyBsYXN0IC1uIDIwIHJlYm9vdDsn',0,0,'最后重启时间'),(16,'查看最近日志','L2Jpbi9zaCAtYyAnZWNobyAwOyB0YWlsIC01MCAvdmFyL2xvZy9tZXNzYWdlczsn',0,0,'查看最近服务器日志'),(17,'查看系统运行服务','L2Jpbi9zaCAtYyAnZWNobyAwO2Noa2NvbmZpZyAtLWxpc3Q7Jw==',0,0,'系统运行服务列表'),(18,'test Top','dG9w',0,0,'Top ');

UNLOCK TABLES;



DROP TABLE IF EXISTS `tag`;

CREATE TABLE `tag` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` char(20) NOT NULL,
  `name_en` char(20) DEFAULT NULL,
  `description` text,
  `alias` char(100) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;


LOCK TABLES `tag` WRITE;

INSERT INTO `tag` VALUES (1,'sds','sdsd','sdfdfsdfgsdf','dsfgdf'),(2,'xcdv','','',''),(3,'xx','zz','zz',NULL);

UNLOCK TABLES;



DROP TABLE IF EXISTS `user`;

CREATE TABLE `user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `email` char(100) NOT NULL,
  `name` char(50) NOT NULL,
  `admin` int(1) NOT NULL DEFAULT '0',
  `password` char(100) NOT NULL,
  `created` int(11) NOT NULL DEFAULT '0',
  `updated` int(11) NOT NULL DEFAULT '0',
  `last_login` int(11) NOT NULL DEFAULT '0',
  `stats` int(11) NOT NULL DEFAULT '0',
  `roleid` int(11) NOT NULL DEFAULT '0',
  `siteid` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`),
  UNIQUE KEY `email` (`email`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8;


LOCK TABLES `user` WRITE;

INSERT INTO `user` VALUES (1,'qeqwq@qq.com','xxx',1,'7c4a8d09ca3762af61e59520943dc26494f8941b',1378826091,1380198982,1378826091,1,1,1),(9,'68812424@qq.com','leiyu',0,'7c4a8d09ca3762af61e59520943dc26494f8941b',1381556978,1393330856,1381556978,1,1,1);

UNLOCK TABLES;



DROP TABLE IF EXISTS `village`;

CREATE TABLE `village` (
  `id` binary(64) NOT NULL,
  `auto_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `body` mediumblob,
  `updated` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `updated` (`updated`),
  KEY `auto_id` (`auto_id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;


LOCK TABLES `village` WRITE;

UNLOCK TABLES;



# 删除留存率 存储过程
DROP PROCEDURE IF EXISTS my_churn_ratio_pr;

# 新建留存率 存储过程
delimiter $$
create procedure  my_churn_ratio_pr(setday int)
begin 

declare yesterday char(8);
declare thisday char(8) ;
declare thatDay char (8) ;
declare signTable varchar(13);
declare tempsql varchar(500);

DROP temporary table if exists that_day;

create temporary table if not exists that_day (
  `uid` char(36),
  `channel` int(2) ,
  index (`uid`)
);
 
truncate TABLE that_day;


DROP temporary table if exists this_day;
create temporary table if not exists this_day(
  `uid` char(36),
  `channel` int(2),
  index (`uid`)
);
 
truncate TABLE this_day;

DROP temporary table if exists that_cache;

create temporary table if not exists that_cache (
  `sum` int,
  `channel` int(2) 
);
 
truncate TABLE that_cache;


set @yesterday = DATE_FORMAT(DATE_SUB(now(), INTERVAL 1 DAY),'%Y%m%d');

#当前日期
set @thisday = @yesterday;


#指定日期
set @thatDay = DATE_FORMAT(DATE_SUB(now(), INTERVAL (setday+1) DAY),'%Y%m%d');

#设置插入表环境
set @signTable = REPLACE ("gm_churn_7day" ,'7' ,setday);


if setday = 1 then 
  set @thisday = DATE_FORMAT(DATE_SUB(now(), INTERVAL 7 DAY),'%Y%m%d');
  set @thatDay = DATE_FORMAT(DATE_SUB(now(), INTERVAL 8 DAY),'%Y%m%d');

elseif setday=7 || setday=30 then
  set @thisDay = DATE_FORMAT(DATE_SUB(now(), INTERVAL (setday+1) DAY),'%Y%m%d');
  set @thatDay = DATE_FORMAT(DATE_SUB(now(), INTERVAL (setday*2+1) DAY),'%Y%m%d');

end if;


insert into that_day (uid,channel) select uid,channel from gm_sign_timeline where `day` >= @thatday and 'day'<@thisday group by uid;

insert into this_day (uid,channel) select uid,channel from gm_sign_timeline where `day` >= @thisday group by uid;

insert into that_cache (sum,channel) select count(uid),channel from that_day where 1 group by channel;

select 
   count(a.uid) as sum, 
   a.channel 
   from 
   that_day a,
   this_day i 
   where 
   a.uid = i.uid 
   group by a.channel;
 
#"INSERT INTO ",@signTable,"(`day`,`channel`,`count`,`churn`,`ratio`)
 #将临时表 与登录缓存表和新用户记录表的指定日期的交集 左连接
SET @tempsql = CONCAT("INSERT INTO ",@signTable,"(`day`,`channel`,`count`,`churn`,`ratio`)  
 select 
 ? as day, 
 d.channel , 
 d.sum as count, 
 if((c.sum is null), d.sum, d.sum-c.sum) as sum,
 if((c.sum/d.sum) is null,1,((d.sum-c.sum)/d.sum)) as ratio 
 from  
 that_cache d  
 left join 
 (select 
   count(a.uid) as sum, 
   a.channel 
   from 
   that_day a,
   this_day i 
   where 
   a.uid = i.uid 
   group by a.channel
 ) c 
 on d.channel = c.channel");

PREPARE mainStmt FROM @tempsql;
#EXECUTE mainStmt using @thisDay;
DEALLOCATE PREPARE mainStmt;

end $$
delimiter ;
-- ----------------------------------------------------


# 开启 event 
SET GLOBAL event_scheduler = 1; 

# 删除留存率 存储过程
DROP PROCEDURE IF EXISTS my_retention_ratio_pr;

# 新建留存率 存储过程
delimiter $$
create procedure  my_retention_ratio_pr(setday int)
begin 

declare thisday char(8) ;
declare thatDay char (8) ;
declare signTable varchar(13);
declare tempsql varchar(500);

#当前日期
set @thisday = DATE_FORMAT(DATE_SUB(now(), INTERVAL 1 DAY),'%Y%m%d');

#指定日期
set @thatDay = DATE_FORMAT(DATE_SUB(now(), INTERVAL (setday+1) DAY),'%Y%m%d');

#设置插入表环境
set @signTable = REPLACE ("gm_sign_7day" ,'7' ,setday);


#创建临时表 用于存储各频道 指定日期的新增用户数
 create temporary table if not exists tmpTable 
 ( 
    sum int , 
    channel int 
 ); 
 truncate TABLE tmpTable;

 insert into tmpTable (sum,channel) select count(uid) ,channel from gm_newuser where 
 timeline = @thatDay group by channel;



 #将临时表 与登录缓存表和新用户记录表的指定日期的交集 左连接
SET @tempsql = CONCAT("INSERT INTO ",@signTable,"(`day`,`channel`,`count`,`retention`,`ratio`)  
 select ? as day,a.channel, a.sum as count, if(b.sum2 is null, 0,b.sum2) as sum, 
 if((b.sum2/a.sum) is null,0,(b.sum2/a.sum)) as ratio from tmpTable a left join (SELECT count(c.uid) 
 as sum2, c.channel FROM `gm_sign_cache` c, `gm_newuser` n WHERE n.timeline =? and 
 n.uid = c.uid and c.day = ? group by c.channel) b on a.channel =b.channel");

PREPARE mainStmt FROM @tempsql;
EXECUTE mainStmt using @thisday, @thatDay, @thisday;
DEALLOCATE PREPARE mainStmt;

end $$
delimiter ;
-- ----------------------------------



DROP PROCEDURE IF EXISTS my_game_user_info_pr;

DELIMITER $$

CREATE  PROCEDURE `my_game_user_info_pr`(
in field varchar(255),
in fvalue varchar(555),
in upstr varchar(555),
in uid varchar(100),
in timeline char(8),
in channel int
)

begin 
declare rels int default 0;
declare tempsql varchar(1500);
declare tempsql2 varchar(1500);

set @uids = uid;
set @channels = channel;
set @timelines = timeline;

 #update guser表
SET @tempsql = CONCAT("INSERT INTO guser (",field,") VALUES (",fvalue,") ON DUPLICATE KEY UPDATE ",upstr);
PREPARE mainStmt FROM @tempsql;
EXECUTE mainStmt ;

select row_count() into rels;
DEALLOCATE PREPARE mainStmt;


if  rels = 1 then
   #插入gm_newuser表
    set @tempsql2 =  CONCAT("INSERT INTO gm_newuser (`uid`,`channel`,`timeline`) values (?,?,?)");
    
  PREPARE mainStmt2 FROM @tempsql2;
  EXECUTE mainStmt2 using @uids, @channels, @timelines;
  DEALLOCATE PREPARE mainStmt2;

end if;

end $$

DELIMITER ;


-- ---------------------------------


delimiter $$
CREATE EVENT IF NOT EXISTS retention_ratio_evt

#每天 00:01:00 执行一次任务
on schedule every 1 DAY starts  '2013-09-01 00:01:00'


ON COMPLETION PRESERVE

DO 
begin
     
     #次日留存
     CALL my_retention_ratio_pr(1);  
     #七日留存     
     CALL my_retention_ratio_pr(7);  
     #30日留存   
     CALL my_retention_ratio_pr(30);

     # 统计日期
     set @yesterday = DATE_FORMAT(DATE_SUB(now(), INTERVAL 1 DAY),'%Y%m%d');
     
     #把缓存写入时间轴
     insert into `gm_sign_timeline` (`dayhh`,`day`,`hh`,`uid`,`channel`) select 
       concat(`day`,`hh`) as `dayhh`,`day`,`hh`,`uid`,`channel` 
       from `gm_sign_cache` where `day` = @yesterday;

     #删除缓存记录 
     SET SQL_SAFE_UPDATES = 0;
     DELETE FROM `gm_sign_cache` WHERE day = @yesterday;

end $$
delimiter ;

# 开启事件
ALTER EVENT retention_ratio_evt ENABLE;

-- --------------------------------------------


delimiter $$
 
CREATE EVENT IF NOT EXISTS churn_ratio_1evt

#每天 00:30:00 执行一次任务
on schedule every 1 DAY starts  '2013-09-01 00:30:00'

ON COMPLETION PRESERVE

DO 
begin
     
     #日流失
     CALL my_churn_ratio_pr(1);  

end $$
delimiter ;

# open Event
ALTER EVENT churn_ratio_1evt ENABLE;

-- -----------------------------------------------


delimiter $$

CREATE EVENT IF NOT EXISTS churn_ratio_7evt

#每周一 02:00:00 执行一次任务
on schedule every 1 WEEK starts  '2013-09-01 02:00:00'


ON COMPLETION PRESERVE

DO 
begin
     
     #周流失
     CALL my_churn_ratio_pr(7);  

end $$
delimiter ;

# open Event
ALTER EVENT churn_ratio_7evt ENABLE;

-- ------------------------------------------------------


delimiter $$
CREATE EVENT IF NOT EXISTS churn_ratio_30evt

#每月 1号 03:00:00 执行一次任务
on schedule every 1 MONTH starts  '2013-09-01 03:00:00'

ON COMPLETION PRESERVE

DO 
begin
     
     #月流失
     CALL my_churn_ratio_pr(30);  

end $$

delimiter ;






# open Event
ALTER EVENT churn_ratio_30evt ENABLE;


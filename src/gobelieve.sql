/*
 Navicat MySQL Data Transfer

 Source Server         : localhost_3306
 Source Server Type    : MySQL
 Source Server Version : 100129
 Source Host           : 127.0.0.1:3306
 Source Schema         : gobelieve

 Target Server Type    : MySQL
 Target Server Version : 100129
 File Encoding         : 65001

 Date: 01/09/2019 13:32:06
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for group
-- ----------------------------
DROP TABLE IF EXISTS `group`;
CREATE TABLE `group` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `appid` bigint(20) DEFAULT NULL,
  `master` bigint(20) DEFAULT NULL,
  `super` tinyint(4) NOT NULL DEFAULT '0',
  `name` varchar(255) DEFAULT NULL,
  `notice` varchar(255) DEFAULT NULL COMMENT '公告',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Table structure for group_member
-- ----------------------------
DROP TABLE IF EXISTS `group_member`;
CREATE TABLE `group_member` (
  `group_id` bigint(20) NOT NULL DEFAULT '0',
  `uid` bigint(20) NOT NULL DEFAULT '0',
  `timestamp` int(11) DEFAULT NULL COMMENT '入群时间,单位：秒',
  `nickname` varchar(255) DEFAULT NULL COMMENT '群内昵称',
  `mute` tinyint(1) DEFAULT '0' COMMENT '群内禁言',
  PRIMARY KEY (`group_id`,`uid`),
  KEY `idx_group_member_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

SET FOREIGN_KEY_CHECKS = 1;

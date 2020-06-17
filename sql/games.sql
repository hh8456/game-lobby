/*
 Navicat Premium Data Transfer

 Source Server         : 内网服务器 dev
 Source Server Type    : MySQL
 Source Server Version : 50730
 Source Host           : 192.168.0.155:3306
 Source Schema         : games

 Target Server Type    : MySQL
 Target Server Version : 50730
 File Encoding         : 65001

 Date: 10/06/2020 09:35:55
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for agent
-- ----------------------------
DROP TABLE IF EXISTS `agent`;
CREATE TABLE `agent`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uid` int(10) NULL DEFAULT NULL,
  `invite_code` int(10) NULL DEFAULT NULL COMMENT '邀请码',
  `is_agent` int(10) NULL DEFAULT NULL COMMENT '是否代理',
  `is_senior_agent` int(10) NULL DEFAULT NULL COMMENT '是否总代',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_uid`(`uid`) USING BTREE,
  INDEX `idx_is_agent`(`is_agent`) USING BTREE,
  INDEX `idx_is_senior_agent`(`is_senior_agent`) USING BTREE,
  INDEX `idx_invite_code`(`invite_code`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for invite_code_in_use
-- ----------------------------
DROP TABLE IF EXISTS `invite_code_in_use`;
CREATE TABLE `invite_code_in_use`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `invite_code` bigint(20) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `invite_code`(`invite_code`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 743 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for niuniu_game_log
-- ----------------------------
DROP TABLE IF EXISTS `niuniu_game_log`;
CREATE TABLE `niuniu_game_log`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `game_id` bigint(20) NOT NULL COMMENT '游戏场次 id',
  `room_type` int(10) NULL DEFAULT NULL COMMENT '游戏类型, 系统房,自建房',
  `room_id` int(10) NULL DEFAULT NULL,
  `player_num` int(10) NULL DEFAULT NULL COMMENT '玩家人数',
  `banker_uid` int(10) NULL DEFAULT NULL COMMENT '庄家 uid',
  `banker_bet_id` bigint(20) NULL DEFAULT NULL,
  `player0` int(10) NULL DEFAULT NULL,
  `player0_bet_id` bigint(20) NULL DEFAULT NULL,
  `player1` int(10) NULL DEFAULT NULL,
  `player1_bet_id` bigint(20) NULL DEFAULT NULL,
  `player2` int(10) NULL DEFAULT NULL,
  `player2_bet_id` bigint(20) NULL DEFAULT NULL,
  `player3` int(10) NULL DEFAULT NULL,
  `player3_bet_id` bigint(20) NULL DEFAULT NULL,
  `player4` int(10) NULL DEFAULT NULL,
  `player4_bet_id` bigint(20) NULL DEFAULT NULL,
  `player5` int(10) NULL DEFAULT NULL,
  `player5_bet_id` bigint(20) NULL DEFAULT NULL,
  `player6` int(10) NULL DEFAULT NULL,
  `player6_bet_id` bigint(20) NULL DEFAULT NULL,
  `player7` int(10) NULL DEFAULT NULL,
  `player7_bet_id` bigint(20) NULL DEFAULT NULL,
  `player8` int(10) NULL DEFAULT NULL,
  `player8_bet_id` bigint(20) NULL DEFAULT NULL,
  `player9` int(10) NULL DEFAULT NULL,
  `player9_bet_id` bigint(20) NULL DEFAULT NULL,
  `room_config` blob NULL COMMENT '当前房间所使用的配置',
  `play_date` datetime(0) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_game_id`(`game_id`) USING BTREE,
  INDEX `idx_blanker_uid_play_date`(`banker_uid`, `play_date`) USING BTREE,
  INDEX `idx_play_date`(`play_date`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for niuniu_person_bet_log
-- ----------------------------
DROP TABLE IF EXISTS `niuniu_person_bet_log`;
CREATE TABLE `niuniu_person_bet_log`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uid` int(10) NULL DEFAULT NULL,
  `bet_id` bigint(20) NULL DEFAULT NULL COMMENT '注局流水号',
  `game_id` bigint(20) NULL DEFAULT NULL COMMENT '游戏场次流水号',
  `room_type` int(10) NULL DEFAULT NULL,
  `room_id` int(10) NULL DEFAULT NULL,
  `is_banker` int(10) NULL DEFAULT NULL COMMENT '0 是闲家, 1 是庄家',
  `bet_rate` int(10) NULL DEFAULT NULL COMMENT '下注倍数',
  `robZhuang_rate` int(10) NULL DEFAULT NULL COMMENT '庄家抢庄倍数',
  `gold_before_change` int(10) NULL DEFAULT NULL COMMENT '变化前的金币',
  `gold_change` int(10) NULL DEFAULT NULL COMMENT '变化的金币',
  `gold_after_change` int(10) NULL DEFAULT NULL COMMENT '变化后的金币',
  `card_type` int(10) NULL DEFAULT NULL COMMENT '牌型, 比如 牛x, 同花顺',
  `card_1` int(10) NULL DEFAULT NULL COMMENT '前面3张牌表示分组,后面两张表示大小',
  `card_2` int(10) NULL DEFAULT NULL,
  `card_3` int(10) NULL DEFAULT NULL,
  `card_4` int(10) NULL DEFAULT NULL,
  `card_5` int(10) NULL DEFAULT NULL,
  `play_date` datetime(0) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_uid_date`(`uid`, `play_date`) USING BTREE,
  INDEX `idx_gameid_uid`(`game_id`, `uid`) USING BTREE,
  INDEX `idx_uid_betid`(`uid`, `bet_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for player_base_info
-- ----------------------------
DROP TABLE IF EXISTS `player_base_info`;
CREATE TABLE `player_base_info`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `wxid` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  `wxid_crc32` bigint(20) UNSIGNED NULL DEFAULT NULL,
  `uid` int(10) NULL DEFAULT NULL,
  `head_pic` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_estonian_ci NULL DEFAULT NULL,
  `invite_code` int(10) NULL DEFAULT NULL,
  `diamond` int(10) NULL DEFAULT NULL,
  `gold` int(10) NULL DEFAULT NULL,
  `sex` tinyint(2) NULL DEFAULT NULL,
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  `reg_date` datetime(0) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_wxid`(`wxid`) USING BTREE,
  INDEX `idx_uid`(`uid`) USING BTREE,
  INDEX `idx_wx_crc32_id`(`wxid_crc32`, `wxid`) USING BTREE,
  INDEX `idx_reg_date`(`reg_date`) USING BTREE,
  INDEX `idx_name`(`name`) USING BTREE,
  INDEX `idx_invite_code`(`invite_code`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1321 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for subordinate
-- ----------------------------
DROP TABLE IF EXISTS `subordinate`;
CREATE TABLE `subordinate`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uid` int(10) NULL DEFAULT NULL COMMENT 'subordinate_uid 的上级',
  `subordinate_uid` int(10) NULL DEFAULT NULL COMMENT 'uid 的直接下级',
  `establish_contact_date` datetime(0) NULL DEFAULT NULL COMMENT 'subordinate 成为 uid 下级的时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_subordinate`(`subordinate_uid`) USING BTREE,
  INDEX `idx_uid_sub_date`(`uid`, `subordinate_uid`, `establish_contact_date`) USING BTREE,
  INDEX `idx_date`(`establish_contact_date`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for uid_in_use
-- ----------------------------
DROP TABLE IF EXISTS `uid_in_use`;
CREATE TABLE `uid_in_use`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `unique_uid`(`uid`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 743 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;

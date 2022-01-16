DROP DATABASE IF EXISTS `webshop`;
CREATE DATABASE IF NOT EXISTS `webshop`;
USE `webshop`;
 
#show warnings; 


####################################################################################
#用户表
#
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` 
(
	`id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,	 
	-- `user_id` INT(10) UNSIGNED NOT NULL COMMENT '用户ID',		#非空约束		
	`user_mobile` VARCHAR(11) NOT NULL COMMENT '用户手机号',		#非空约束		
	`user_password` VARCHAR(32) NOT NULL COMMENT '用户密码',
	`nickname` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '用户昵称',	#默认值约束，默认值约束通常用在已经设置了非空约束的列，这样能够防止数据表在录入数据时出现错误。	
	`role` INT(11) NOT NULL DEFAULT 0 COMMENT '用户角色: 0,普通用户 1,管理员',
	`head_url` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '头像URL',
	`gender` ENUM ('male','female') NULL COMMENT '性别: male,男 female,女',
	-- `age` TINYINT NULL COMMENT '年龄' CHECK(age>=0 AND age<=200),	#基于列的 CHECK 约束，将 CHECK 约束子句置于表中某个列的定义之后
	`birthday` date DEFAULT NULL COMMENT '生日',		
	#`address` varchar(255) DEFAULT NULL COMMENT '地址',
	`desc` varchar(255) DEFAULT NULL COMMENT '个人简介',	# TextField
	
	`is_deleted` TINYINT(1) UNSIGNED DEFAULT 0 COMMENT '是否删除：0为false, 非0为真',
	`update_time` timestamp NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
	`add_time` timestamp NOT NULL COMMENT '添加时间', 
	PRIMARY KEY (`id`),
	UNIQUE KEY `unik_user_mobile` (`user_mobile`)			#唯一约束 
);
#desc  `user`;	
#INSERT  `user` (user_mobile, user_password,nickname,add_time) VALUES ('13800138000', '123456', '小小', NOW());
#SELECT * from `user`;
#UPDATE `user` SET user_password = '111' WHERE user_mobile = '13800138000'

####################################################################################
#商品类别表
#
DROP TABLE IF EXISTS `category`;
CREATE TABLE `category` 
(
	`id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
	`name` VARCHAR(16) NOT NULL COMMENT '类别名称',
	`parent_category_id` INT(11) UNSIGNED DEFAULT 0 COMMENT '父类别',	#自外键： 一级类别可以没有父类别  
	`level` INT(1) NOT NULL DEFAULT 1 COMMENT '级别',
	`is_tab` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否显示在首页tab',
	
	`is_deleted` TINYINT(1) UNSIGNED DEFAULT 0 COMMENT '是否删除：0为false, 非0为真',
	`update_time` timestamp NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
	`add_time` timestamp NOT NULL COMMENT '添加时间', 
	PRIMARY KEY (`id`)
	#CONSTRAINT `fk_category_id` FOREIGN KEY(`parent_category_id`) REFERENCES `category`(`id`) 		#外键约束
);
#alter TABLE category ADD FOREIGN KEY(`parent_category_id`) REFERENCES category(`id`);

INSERT  `category` (name, parent_category_id, level, add_time) VALUES ('数据产品', 0, 1, NOW());
INSERT  `category` (name, parent_category_id, level, add_time) VALUES ('手机', 1, 2, NOW());
INSERT  `category` (name, parent_category_id, level, add_time) VALUES ('电脑', 1, 2, NOW());
INSERT  `category` (name, parent_category_id, level, add_time) VALUES ('相机', 1, 2, NOW());
INSERT  `category` (name, parent_category_id, level, add_time) VALUES ('单反相机', 2, 3, NOW());
INSERT  `category` (name, parent_category_id, level, add_time) VALUES ('普通相机', 2, 3, NOW());
INSERT  `category` (name, parent_category_id, level, add_time) VALUES ('台式电脑', 3, 3, NOW());
  
#select id, name, parent_category_id,level,is_tab  from category where is_deleted = 0


####################################################################################
#商品品牌表 
#
DROP TABLE IF EXISTS `brands`;
CREATE TABLE `brands` 
(
	#`id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
	`name` VARCHAR(16) NOT NULL COMMENT '品牌名称', 
	`logo` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '图标地址',
	`is_deleted` TINYINT(1) UNSIGNED DEFAULT 0 COMMENT '是否删除：0为false, 非0为真',
	`update_time` timestamp NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
	`add_time` timestamp NOT NULL COMMENT '添加时间', 
	#PRIMARY KEY (`id`)
	UNIQUE KEY `unik_name` (`name`)			#唯一约束 
);

####################################################################################
#商品表， 分布式的事务最好的解决方案 就是不要让分布式事务出现
#
DROP TABLE IF EXISTS `goods`;
CREATE TABLE `goods` 
(
	`id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
	`category_id` INT(11) UNSIGNED NOT NULL COMMENT '商品类别ID',	#自外键： 一级类别可以没有父类别  
	`brands_name` VARCHAR(16) NOT NULL COMMENT '品牌名称',		#自外键： 一级类别可以没有父类别  
		
	`on_sale` TINYINT(1) UNSIGNED DEFAULT 1 COMMENT '是否上架：0为false, 非0为真',
	`goods_sn` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '商品唯一货号',
	`name` VARCHAR(16) NOT NULL COMMENT '商品名称',
	`click_num` INT(11) UNSIGNED DEFAULT 0 COMMENT '点击数',
	`sold_num` INT(11) UNSIGNED DEFAULT 0 COMMENT '商品销售量',
	`fav_num` INT(11) UNSIGNED DEFAULT 0 COMMENT '收藏数',	#库存是电商中一个重要的环节
	`market_price` FLOAT8 DEFAULT 0.00 COMMENT '市场价格',
	`shop_price` FLOAT8 DEFAULT 0.00 COMMENT '本店价格',
	 
	`is_ship_free` TINYINT(1) UNSIGNED DEFAULT 0 COMMENT '是否免运费：0为false, 非0为真',
	`is_new` TINYINT(1) UNSIGNED DEFAULT 0 COMMENT '是否新品：0为false, 非0为真',
	`is_hot` TINYINT(1) UNSIGNED DEFAULT 0 COMMENT '是否热销：0为false, 非0为真',
	
	`goods_brief` VARCHAR(128) DEFAULT '' COMMENT '商品简短描述',
	`goods_front_image` VARCHAR(128) DEFAULT '' COMMENT '商品封面图',
	`josn_images` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '商品轮播图(Json格式)',	# JSON格式数据
	`desc_images` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '详情页图片(Json格式)',	# JSON格式数据
		
	`is_deleted` TINYINT(1) UNSIGNED DEFAULT 0 COMMENT '是否删除：0为false, 非0为真',
	`update_time` timestamp NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
	`add_time` timestamp NOT NULL COMMENT '添加时间', 
	PRIMARY KEY (`id`)
	#INDEX KEY `idx_category_id`(`category_id`),		#创建普通索引
	
);
 
####################################################################################
#品牌分类表  品牌-品类
# 
DROP TABLE IF EXISTS `BrandsCategory`;
CREATE TABLE `BrandsCategory` 
(
	`id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
	`category_id` INT(11) UNSIGNED NOT NULL COMMENT '商品类别ID',	#自外键： 一级类别可以没有父类别  
	`brands_name` VARCHAR(16) NOT NULL COMMENT '品牌名称',		#自外键： 一级类别可以没有父类别  
		 
	`is_deleted` TINYINT(1) UNSIGNED DEFAULT 0 COMMENT '是否删除：0为false, 非0为真',
	`update_time` timestamp NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
	`add_time` timestamp NOT NULL COMMENT '添加时间', 
	PRIMARY KEY (`id`),
	UNIQUE KEY `unik_category_id_brands_name` (`category_id`, `brands_name`) #唯一索引（联合）
); 

####################################################################################
#轮播的商品表 
#
DROP TABLE IF EXISTS `banner`;
CREATE TABLE `banner` 
(
	`id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
	`image` VARCHAR(16) NOT NULL COMMENT '图片url', 
	`url` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '访问url',
	`index` INT(11) UNSIGNED DEFAULT 0 COMMENT '轮播顺序',
	
	`update_time` timestamp NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
	`add_time` timestamp NOT NULL COMMENT '添加时间', 
	PRIMARY KEY (`id`) 
);



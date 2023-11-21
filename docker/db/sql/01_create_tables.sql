-- Table for tasks
DROP TABLE IF EXISTS `tasks`;

CREATE TABLE `tasks` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `title` varchar(50) NOT NULL,
    `is_done` boolean NOT NULL DEFAULT b'0',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- 締切の追加 (デフォルトは現在時刻)
    `deadline` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- 優先度の追加 (0:低, 1:中, 2:高)
    `priority` int(1) NOT NULL DEFAULT 0,
    `overview` varchar(256)  NOT NULL DEFAULT '-',
    `tag` varchar(50) DEFAULT "null",
    PRIMARY KEY (`id`)
) DEFAULT CHARSET=utf8mb4;

-- Table for users
DROP TABLE IF EXISTS `users`;

CREATE TABLE `users` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `name` varchar(50) NOT NULL UNIQUE,
    `password` binary(32) NOT NULL,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `ownership`;
 
CREATE TABLE `ownership` (
    `user_id` bigint(20) NOT NULL,
    `task_id` bigint(20) NOT NULL,
    PRIMARY KEY (`user_id`, `task_id`)
) DEFAULT CHARSET=utf8mb4;
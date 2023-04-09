-- CREATE DATABASE chat_app;

USE chat_app;

CREATE TABLE users (
    user_id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL
);

CREATE TABLE chat_logs (
    chat_id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    message TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- CREATE TABLE blog_posts (
--     post_id INT PRIMARY KEY AUTO_INCREMENT,
--     user_id INT NOT NULL,
--     title VARCHAR(255) NOT NULL,
--     content TEXT NOT NULL,
--     image_path VARCHAR(255),
--     video_path VARCHAR(255),
--     timestamp DATETIME NOT NULL,
--     FOREIGN KEY (user_id) REFERENCES users(user_id)
-- );
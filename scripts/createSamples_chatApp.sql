USE chat_app;


INSERT INTO users (username, password)
VALUES ('admin', 'admin'),
	   ('john_doe', 'password123'),
       ('jane_doe', 'password456');

INSERT INTO chat_logs (user_id, message, timestamp)
VALUES (2, 'Hello, Jane!', NOW()),
       (3, 'Hi, John!', NOW()),
       (2, 'How are you?', NOW()),
       (3, 'I am doing well, thanks!', NOW());

-- INSERT INTO blog_posts (user_id, title, content, image_path, video_path, timestamp)
-- VALUES (2, 'My first blog post', 'Lorem ipsum dolor sit amet, consectetur adipiscing elit.', 'image1.jpg', NULL, NOW()),
--        (3, 'My second blog post', 'Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.', NULL, 'video1.mp4', NOW());
       
       
       

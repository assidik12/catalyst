-- Delete all products first (foreign key constraint)
DELETE FROM `products`;

-- Delete all categories
DELETE FROM `category`;

-- Reset auto-increment (optional)
ALTER TABLE `products` AUTO_INCREMENT = 1;
ALTER TABLE `category` AUTO_INCREMENT = 1;

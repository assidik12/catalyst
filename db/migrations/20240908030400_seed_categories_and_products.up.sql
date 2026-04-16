-- Insert Categories
INSERT INTO `category` (`name`, `stock`) VALUES
('Electronics', 500),
('Fashion', 800),
('Home & Living', 600),
('Sports & Outdoor', 400),
('Books & Stationery', 350),
('Beauty & Health', 450),
('Toys & Games', 300),
('Automotive', 250),
('Food & Beverages', 700);

-- Insert Products for Electronics (Category ID: 1)
INSERT INTO `products` (`name`, `price`, `stock`, `img`, `description`, `category_id`) VALUES
('Samsung Galaxy S23 Ultra', 15999000, 50, 'https://images.unsplash.com/photo-1610945415295-d9bbf067e59c', 'Flagship smartphone with 200MP camera', 1),
('MacBook Pro M3 14"', 28999000, 30, 'https://images.unsplash.com/photo-1517336714731-489689fd1ca8', 'Powerful laptop for professionals', 1),
('Sony WH-1000XM5', 4999000, 80, 'https://images.unsplash.com/photo-1546435770-a3e426bf472b', 'Premium noise-cancelling headphones', 1),
('iPad Air 5th Gen', 9999000, 45, 'https://images.unsplash.com/photo-1544244015-0df4b3ffc6b0', 'Versatile tablet for work and play', 1),
('Logitech MX Master 3S', 1499000, 120, 'https://images.unsplash.com/photo-1527814050087-3793815479db', 'Ergonomic wireless mouse', 1),
('Dell UltraSharp 27" 4K', 5999000, 40, 'https://images.unsplash.com/photo-1527443224154-c4a3942d3acf', 'Professional 4K monitor', 1),

-- Insert Products for Fashion (Category ID: 2)
('Nike Air Max 270', 1899000, 150, 'https://images.unsplash.com/photo-1542291026-7eec264c27ff', 'Comfortable running shoes', 2),
('Levi\'s 501 Original Jeans', 899000, 200, 'https://images.unsplash.com/photo-1542272604-787c3835535d', 'Classic straight fit denim', 2),
('Adidas Ultraboost 22', 2499000, 100, 'https://images.unsplash.com/photo-1608231387042-66d1773070a5', 'Performance running shoes', 2),
('Zara Cotton T-Shirt', 299000, 300, 'https://images.unsplash.com/photo-1521572163474-6864f9cf17ab', 'Basic cotton tee for daily wear', 2),
('H&M Slim Fit Chinos', 499000, 180, 'https://images.unsplash.com/photo-1473966968600-fa801b869a1a', 'Versatile casual pants', 2),
('Fossil Gen 6 Smartwatch', 3499000, 60, 'https://images.unsplash.com/photo-1523275335684-37898b6baf30', 'Stylish smartwatch with health tracking', 2),

-- Insert Products for Home & Living (Category ID: 3)
('IKEA POÄNG Armchair', 1299000, 80, 'https://images.unsplash.com/photo-1555041469-a586c61ea9bc', 'Comfortable bent wood armchair', 3),
('Philips Hue Smart Bulb Set', 899000, 150, 'https://images.unsplash.com/photo-1558618666-fcd25c85cd64', 'RGB smart lighting system', 3),
('Dyson V15 Detect', 8999000, 35, 'https://images.unsplash.com/photo-1558317374-067fb5f30001', 'Cordless vacuum with laser detection', 3),
('Nespresso Vertuo Next', 2499000, 70, 'https://images.unsplash.com/photo-1517668808822-9ebb02f2a0e6', 'Premium coffee maker', 3),
('Muji Cotton Bedsheet Set', 599000, 200, 'https://images.unsplash.com/photo-1631049307264-da0ec9d70304', 'Soft and breathable bedding', 3),
('Xiaomi Air Purifier 4', 2299000, 90, 'https://images.unsplash.com/photo-1585771724684-38269d6639fd', 'Smart air purifier with HEPA filter', 3),

-- Insert Products for Sports & Outdoor (Category ID: 4)
('Yoga Mat Premium', 399000, 250, 'https://images.unsplash.com/photo-1601925260368-ae2f83cf8b7f', 'Non-slip exercise mat 6mm thick', 4),
('Decathlon Mountain Bike', 4999000, 40, 'https://images.unsplash.com/photo-1576435728678-68d0fbf94e91', '21-speed mountain bicycle', 4),
('The North Face Backpack', 1899000, 120, 'https://images.unsplash.com/photo-1553062407-98eeb64c6a62', 'Durable hiking backpack 40L', 4),
('Garmin Forerunner 255', 5499000, 55, 'https://images.unsplash.com/photo-1575311373937-040b8e1fd5b6', 'GPS running watch with training metrics', 4),
('Wilson Evolution Basketball', 899000, 80, 'https://images.unsplash.com/photo-1546519638-68e109498ffc', 'Official size composite leather ball', 4),

-- Insert Products for Books & Stationery (Category ID: 5)
('Atomic Habits - James Clear', 189000, 300, 'https://images.unsplash.com/photo-1544947950-fa07a98d237f', 'Bestselling self-improvement book', 5),
('Moleskine Classic Notebook', 249000, 400, 'https://images.unsplash.com/photo-1531346878377-a5be20888e57', 'Premium hardcover notebook A5', 5),
('Faber-Castell Pencil Set', 159000, 500, 'https://images.unsplash.com/photo-1513542789411-b6a5d4f31634', 'Professional drawing pencils 12pcs', 5),
('The Psychology of Money', 179000, 250, 'https://images.unsplash.com/photo-1592496431122-2349e0fbc666', 'Finance and investing book', 5),
('Staedtler Marker Set', 299000, 350, 'https://images.unsplash.com/photo-1513542789411-b6a5d4f31634', 'Permanent markers 24 colors', 5),

-- Insert Products for Beauty & Health (Category ID: 6)
('Cetaphil Gentle Cleanser', 189000, 400, 'https://images.unsplash.com/photo-1556228720-195a672e8a03', 'Dermatologist recommended face wash', 6),
('The Ordinary Niacinamide', 119000, 500, 'https://images.unsplash.com/photo-1620916566398-39f1143ab7be', 'Serum for blemish-prone skin', 6),
('Dove Body Wash', 89000, 600, 'https://images.unsplash.com/photo-1535585209827-a15fcdbc4c2d', 'Moisturizing body cleanser', 6),
('Maybelline Fit Me Foundation', 149000, 350, 'https://images.unsplash.com/photo-1522335789203-aabd1fc54bc9', 'Natural finish liquid foundation', 6),
('Omron Blood Pressure Monitor', 899000, 80, 'https://images.unsplash.com/photo-1584308666744-24d5c474f2ae', 'Digital BP monitor with memory', 6),

-- Insert Products for Toys & Games (Category ID: 7)
('LEGO Star Wars Millennium Falcon', 1499000, 60, 'https://images.unsplash.com/photo-1587654780291-39c9404d746b', 'Iconic spaceship building set 1351pcs', 7),
('Nintendo Switch OLED', 4999000, 45, 'https://images.unsplash.com/photo-1578303512597-81e6cc155b3e', 'Handheld gaming console', 7),
('Rubik\'s Cube Speed Cube', 149000, 300, 'https://images.unsplash.com/photo-1591991731833-b8b1e01c0d72', 'Smooth turning puzzle cube', 7),
('Hot Wheels Track Set', 499000, 150, 'https://images.unsplash.com/photo-1558060370-d644479cb6f7', 'Loop racing track with 2 cars', 7),
('Monopoly Classic Edition', 299000, 200, 'https://images.unsplash.com/photo-1611891487953-bc36de7d0e0e', 'Family board game', 7),

-- Insert Products for Automotive (Category ID: 8)
('Michelin Pilot Sport 4', 1899000, 100, 'https://images.unsplash.com/photo-1486262715619-67b85e0b08d3', 'High performance tire 225/45R17', 8),
('Bosch Car Battery 55Ah', 1299000, 120, 'https://images.unsplash.com/photo-1619642751034-765dfdf7c58e', 'Maintenance-free automotive battery', 8),
('3M Car Wax Ultimate', 249000, 200, 'https://images.unsplash.com/photo-1607860108855-64acf2078ed9', 'Premium synthetic wax protection', 8),
('Garmin Dash Cam 66W', 3499000, 50, 'https://images.unsplash.com/photo-1563720360172-67b8f3dce741', 'Wide angle HD dashcam', 8),

-- Insert Products for Food & Beverages (Category ID: 9)
('Lavazza Espresso Beans 1kg', 299000, 400, 'https://images.unsplash.com/photo-1559056199-641a0ac8b55e', 'Premium Italian coffee beans', 9),
('Haribo Goldbears Gummy', 45000, 800, 'https://images.unsplash.com/photo-1582058091505-f87a2e55a40f', 'Classic fruit gummy bears 200g', 9),
('Pringles Original', 35000, 1000, 'https://images.unsplash.com/photo-1621939514649-280e2ee25f60', 'Stackable potato crisps 107g', 9),
('Red Bull Energy Drink', 25000, 1200, 'https://images.unsplash.com/photo-1622543925917-763c34c1a885', 'Energy drink 250ml can', 9),
('Ferrero Rocher T16', 159000, 500, 'https://images.unsplash.com/photo-1548848149-e5addb62e185', 'Premium chocolate pralines', 9);

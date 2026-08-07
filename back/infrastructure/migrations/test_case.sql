-- Цепочка обмена для получения RTX4080:
-- RTX4080 -> Ноутбук -> Велосипед -> Телефон
-- Алгоритм поиска стартует с желаемого товара
-- и идёт к товару пользователя через wishlist владельцев
--
-- Пароль всех четырёх демо-пользователей: demo1234
-- В password_hash лежит именно bcrypt-хеш: /auth/login сверяет пароль через
-- bcrypt.CompareHashAndPassword, и с произвольной строкой вход невозможен.
-- ============================================================
TRUNCATE TABLE customers CASCADE;
DO $$
DECLARE
cust1 UUID;
cust2 UUID;
cust3 UUID;
cust4 UUID;

cat_phone UUID;
cat_bike UUID;
cat_laptop UUID;
cat_gpu UUID;

prod_phone UUID;
prod_bike UUID;
prod_laptop UUID;
prod_gpu UUID;

wish_phone UUID;
wish_bike UUID;
wish_laptop UUID;
wish_gpu UUID;
BEGIN

-- Пользователи
INSERT INTO customers(email,password_hash)
VALUES ('user1@test.com','$2a$10$d1l2vg5i67/BH59rklxg/uuTkM4dUsgHmw8EZXBkGTJYK9ixKySna')
RETURNING customer_id INTO cust1;

INSERT INTO customers(email,password_hash)
VALUES ('user2@test.com','$2a$10$d1l2vg5i67/BH59rklxg/uuTkM4dUsgHmw8EZXBkGTJYK9ixKySna')
RETURNING customer_id INTO cust2;

INSERT INTO customers(email,password_hash)
VALUES ('user3@test.com','$2a$10$d1l2vg5i67/BH59rklxg/uuTkM4dUsgHmw8EZXBkGTJYK9ixKySna')
RETURNING customer_id INTO cust3;

INSERT INTO customers(email,password_hash)
VALUES ('user4@test.com','$2a$10$d1l2vg5i67/BH59rklxg/uuTkM4dUsgHmw8EZXBkGTJYK9ixKySna')
RETURNING customer_id INTO cust4;


-- Категории
INSERT INTO categories(name)
VALUES ('Телефон')
RETURNING category_id INTO cat_phone;

INSERT INTO categories(name)
VALUES ('Велосипед')
RETURNING category_id INTO cat_bike;

INSERT INTO categories(name)
VALUES ('Ноутбук')
RETURNING category_id INTO cat_laptop;

INSERT INTO categories(name)
VALUES ('Видеокарта')
RETURNING category_id INTO cat_gpu;


-- Товары пользователей
-- User1 имеет телефон
INSERT INTO products(customer_id,category_id,title,description)
VALUES
(cust1,cat_phone,'iPhone 15','Телефон пользователя')
RETURNING product_id INTO prod_phone;

-- User2 имеет велосипед
INSERT INTO products(customer_id,category_id,title,description)
VALUES
(cust2,cat_bike,'GT Avalanche','Велосипед пользователя')
RETURNING product_id INTO prod_bike;

-- User3 имеет ноутбук
INSERT INTO products(customer_id,category_id,title,description)
VALUES
(cust3,cat_laptop,'MacBook Pro','Ноутбук пользователя')
RETURNING product_id INTO prod_laptop;

-- User4 имеет видеокарту
INSERT INTO products(customer_id,category_id,title,description)
VALUES
(cust4,cat_gpu,'RTX 4080','Видеокарта пользователя')
RETURNING product_id INTO prod_gpu;


-- Wishlist каждого объявления:
-- владелец объявления указывает, что он хочет получить взамен

INSERT INTO wishlists(product_id,name)
VALUES
(prod_phone,'Хочу видеокарту')
RETURNING wishlist_id INTO wish_phone;

INSERT INTO wishlists(product_id,name)
VALUES
(prod_bike,'Хочу телефон')
RETURNING wishlist_id INTO wish_bike;

INSERT INTO wishlists(product_id,name)
VALUES
(prod_laptop,'Хочу велосипед')
RETURNING wishlist_id INTO wish_laptop;

INSERT INTO wishlists(product_id,name)
VALUES
(prod_gpu,'Хочу ноутбук')
RETURNING wishlist_id INTO wish_gpu;


-- Категории wishlist:
-- User2 отдаёт велосипед и хочет телефон
INSERT INTO wishlist_options
VALUES
(wish_bike,cat_phone);

-- User3 отдаёт ноутбук и хочет велосипед
INSERT INTO wishlist_options
VALUES
(wish_laptop,cat_bike);

-- User4 отдаёт видеокарту и хочет ноутбук
INSERT INTO wishlist_options
VALUES
(wish_gpu,cat_laptop);

-- User1 отдаёт телефон и хочет видеокарту
-- (конечная точка цепочки для поиска)
INSERT INTO wishlist_options
VALUES
(wish_phone,cat_gpu);


-- Отзывы (не участвуют в поиске, но пригодятся позже)
INSERT INTO reviews(from_customer_id,to_customer_id,product_id,rating)
VALUES
(cust1,cust2,prod_bike,5),
(cust2,cust3,prod_laptop,5),
(cust3,cust4,prod_gpu,5);


RAISE NOTICE '';
RAISE NOTICE '==========================================';
RAISE NOTICE 'Тестовая база создана';
RAISE NOTICE '==========================================';
RAISE NOTICE 'User1 имеет:      iPhone';
RAISE NOTICE 'User2 имеет:      Велосипед';
RAISE NOTICE 'User3 имеет:      Ноутбук';
RAISE NOTICE 'User4 имеет:      RTX4080';
RAISE NOTICE '';
RAISE NOTICE 'User1 хочет получить RTX4080';
RAISE NOTICE 'Искомая цепочка:';
RAISE NOTICE 'RTX4080 -> MacBook Pro -> GT Avalanche -> iPhone 15';
RAISE NOTICE '==========================================';

END;
$$;
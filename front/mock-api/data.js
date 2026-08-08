const products = [
    {
        product_id: 'avito-gpu-rtx-3060',
        customer_id: 'user-pskov-01',
        category_id: 'video-cards',
        category: 'Видеокарты',
        name: 'Видеокарта GeForce RTX 3060 12 ГБ',
        description: 'В отличном состоянии, использовалась для игр. Полный комплект, проверка при встрече.',
        price: 28990,
        currency: 'RUB',
        location: 'Псков',
        image_url: 'https://50.img.avito.st/image/1/1.iedNrLa4JQ57G6cDU56eolkMJwjzDacYewAnDP0FLQT7._xzbVJNce46KNeP-4N3tbbh2TTVb5eNgmHDUYIyRsnU',
        status: 'active',
        created_at: '2026-08-06T18:40:00Z',
        updated_at: '2026-08-06T18:40:00Z',
    },
    {
        product_id: 'avito-gpu-rx-6600',
        customer_id: 'user-pskov-02',
        category_id: 'video-cards',
        category: 'Видеокарты',
        name: 'AMD Radeon RX 6600 8GB',
        description: 'Тихая игровая видеокарта, не майнилась. Возможен обмен на консоль или игры.',
        price: 21500,
        currency: 'RUB',
        location: 'Псков',
        image_url: 'https://images.unsplash.com/photo-1591488320449-011701bb6704?auto=format&fit=crop&w=900&q=80',
        status: 'active',
        created_at: '2026-08-05T12:15:00Z',
        updated_at: '2026-08-05T12:15:00Z',
    },
    {
        product_id: 'avito-game-ps5-spider-man',
        customer_id: 'user-pskov-03',
        category_id: 'console-games',
        category: 'Игры для приставок',
        name: "Marvel's Spider-Man 2 для PS5",
        description: 'Физический диск, оригинальное издание. Диск и коробка без повреждений.',
        price: 3990,
        currency: 'RUB',
        location: 'Псков',
        image_url: 'https://images.unsplash.com/photo-1605899435973-ca2d1a8861cf?auto=format&fit=crop&w=900&q=80',
        status: 'active',
        created_at: '2026-08-07T09:30:00Z',
        updated_at: '2026-08-07T09:30:00Z',
    },
    {
        product_id: 'avito-game-ps4-rdr2',
        customer_id: 'user-pskov-04',
        category_id: 'console-games',
        category: 'Игры для приставок',
        name: 'Red Dead Redemption 2 для PS4',
        description: 'Физическое издание на русском языке, состояние отличное.',
        price: 2490,
        currency: 'RUB',
        location: 'Псков',
        image_url: 'https://images.unsplash.com/photo-1542751371-adc38448a05e?auto=format&fit=crop&w=900&q=80',
        status: 'active',
        created_at: '2026-08-04T16:05:00Z',
        updated_at: '2026-08-04T16:05:00Z',
    },
    {
        product_id: 'avito-game-xbox-forza',
        customer_id: 'user-pskov-05',
        category_id: 'console-games',
        category: 'Игры для приставок',
        name: 'Forza Horizon 5 для Xbox',
        description: 'Лицензионный диск, полностью рабочий. Обмен на игры для PS5.',
        price: 2990,
        currency: 'RUB',
        location: 'Псков',
        image_url: 'https://images.unsplash.com/photo-1605901309584-818e25960a8f?auto=format&fit=crop&w=900&q=80',
        status: 'active',
        created_at: '2026-08-03T11:20:00Z',
        updated_at: '2026-08-03T11:20:00Z',
    },
    {
        product_id: 'avito-gpu-rtx-3070',
        customer_id: 'user-pskov-06',
        category_id: 'video-cards',
        category: 'Видеокарты',
        name: 'GeForce RTX 3070 Gaming OC 8 ГБ',
        description: 'Рабочая видеокарта для игр в 2K. Проверка и самовывоз в Пскове.',
        price: 32990,
        currency: 'RUB',
        location: 'Псков',
        image_url: 'https://images.unsplash.com/photo-1591488320449-011701bb6704?auto=format&fit=crop&w=900&q=80',
        status: 'active',
        created_at: '2026-08-02T14:10:00Z',
        updated_at: '2026-08-02T14:10:00Z',
    },
    {
        product_id: 'avito-gpu-gtx-1660',
        customer_id: 'user-pskov-07',
        category_id: 'video-cards',
        category: 'Видеокарты',
        name: 'Видеокарта GTX 1660 Super',
        description: 'Аккуратное состояние, работает стабильно. Возможен обмен.',
        price: 16900,
        currency: 'RUB',
        location: 'Псков',
        image_url: '',
        status: 'active',
        created_at: '2026-08-01T10:45:00Z',
        updated_at: '2026-08-01T10:45:00Z',
    },
    {
        product_id: 'avito-game-ps5-god-of-war',
        customer_id: 'user-pskov-08',
        category_id: 'console-games',
        category: 'Игры для приставок',
        name: 'God of War Ragnarök для PS5',
        description: 'Физический диск, коробка без сколов. Можно обменять на другую игру.',
        price: 4490,
        currency: 'RUB',
        location: 'Псков',
        image_url: 'https://images.unsplash.com/photo-1605899435973-ca2d1a8861cf?auto=format&fit=crop&w=900&q=80',
        status: 'active',
        created_at: '2026-07-31T19:20:00Z',
        updated_at: '2026-07-31T19:20:00Z',
    },
    {
        product_id: 'avito-game-ps4-gta-v',
        customer_id: 'user-pskov-09',
        category_id: 'console-games',
        category: 'Игры для приставок',
        name: 'GTA V для PS4',
        description: 'Диск полностью рабочий, есть оригинальная коробка.',
        price: 1990,
        currency: 'RUB',
        location: 'Псков',
        image_url: '',
        status: 'active',
        created_at: '2026-07-30T15:00:00Z',
        updated_at: '2026-07-30T15:00:00Z',
    },
    {
        product_id: 'avito-game-ps5-hogwarts',
        customer_id: 'user-pskov-10',
        category_id: 'console-games',
        category: 'Игры для приставок',
        name: 'Hogwarts Legacy для PS5',
        description: 'Физическое издание в хорошем состоянии, один владелец.',
        price: 3590,
        currency: 'RUB',
        location: 'Псков',
        image_url: 'https://images.unsplash.com/photo-1542751371-adc38448a05e?auto=format&fit=crop&w=900&q=80',
        status: 'active',
        created_at: '2026-07-29T11:35:00Z',
        updated_at: '2026-07-29T11:35:00Z',
    },
    {
        product_id: 'avito-game-xbox-halo',
        customer_id: 'user-pskov-11',
        category_id: 'console-games',
        category: 'Игры для приставок',
        name: 'Halo Infinite для Xbox',
        description: 'Диск в хорошем состоянии. Рассмотрю обмен на игры для Xbox.',
        price: 2290,
        currency: 'RUB',
        location: 'Псков',
        image_url: '',
        status: 'active',
        created_at: '2026-07-28T09:15:00Z',
        updated_at: '2026-07-28T09:15:00Z',
    },
];

const categories = [
    { category_id: 'computer-goods', name: 'Товары для компьютера' },
    { category_id: 'components', name: 'Комплектующие', parent_id: 'computer-goods' },
    { category_id: 'video-cards', name: 'Видеокарты', parent_id: 'components' },
    { category_id: 'console-games', name: 'Игры для приставок' },
];

const customers = [
    { customer_id: 'user-pskov-01', email: 'alexey@example.com', password: 'password123', is_active: true },
    { customer_id: 'user-pskov-02', email: 'maria@example.com', password: 'password123', is_active: true },
    { customer_id: 'user-pskov-03', email: 'ivan@example.com', password: 'password123', is_active: true },
    { customer_id: 'user-pskov-04', email: 'olga@example.com', password: 'password123', is_active: true },
];

const chains = [
    {
        chain_id: 'chain-pskov-01',
        from_product_id: 'avito-gpu-rtx-3060',
        to_product_id: 'avito-game-ps5-spider-man',
        initiator_id: 'user-pskov-01',
        status: 'active',
        message: 'Готов обменять видеокарту на комплект игр для PS5.',
        created_at: '2026-08-06T19:00:00Z',
        updated_at: '2026-08-06T19:00:00Z',
    },
    {
        chain_id: 'chain-pskov-02',
        from_product_id: 'avito-game-ps4-rdr2',
        to_product_id: 'avito-gpu-rx-6600',
        initiator_id: 'user-pskov-04',
        status: 'pending',
        message: 'Рассмотрю обмен с доплатой.',
        created_at: '2026-08-05T13:00:00Z',
        updated_at: '2026-08-05T13:00:00Z',
    },
];

const reviews = [
    {
        review_id: 'review-pskov-01',
        from_customer_id: 'user-pskov-02',
        to_customer_id: 'user-pskov-01',
        product_id: 'avito-gpu-rtx-3060',
        rating: 5,
        comment: 'Быстро договорились, товар соответствует описанию.',
        created_at: '2026-08-01T10:00:00Z',
        updated_at: '2026-08-01T10:00:00Z',
    },
    {
        review_id: 'review-pskov-02',
        from_customer_id: 'user-pskov-03',
        to_customer_id: 'user-pskov-01',
        product_id: 'avito-gpu-rtx-3060',
        rating: 4,
        comment: 'Всё хорошо, встретились в удобном месте.',
        created_at: '2026-08-02T14:00:00Z',
        updated_at: '2026-08-02T14:00:00Z',
    },
];

const wishlists = [
    {
        wishlist_id: 'wishlist-pskov-01',
        product_id: 'avito-game-ps5-spider-man',
        name: 'Что хочу получить за видеокарту',
        created_at: '2026-08-06T18:50:00Z',
        updated_at: '2026-08-06T18:50:00Z',
    },
];

const wishlistOptions = {
    'wishlist-pskov-01': ['console-games'],
};

const mockNow = '2026-08-07T00:00:00Z';

products.splice(0, products.length, ...products.map((product) => ({
    product_id: product.product_id,
    customer_id: product.customer_id,
    ...(product.category_id ? { category_id: product.category_id } : {}),
    name: product.name,
    ...(product.description ? { description: product.description } : {}),
    is_active: product.status === 'active',
    created_at: product.created_at,
    updated_at: product.updated_at,
})));

categories.forEach((category) => {
    category.created_at = mockNow;
    category.updated_at = mockNow;
});

customers.forEach((customer) => {
    const now = mockNow;
    customer.created_at = now;
    customer.updated_at = now;
});

export { categories, customers, products, chains, reviews, wishlists, wishlistOptions };

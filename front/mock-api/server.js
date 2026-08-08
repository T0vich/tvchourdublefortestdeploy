/* global URL, console, process */

import { createServer } from 'node:http';
import { randomUUID } from 'node:crypto';
import { categories, customers, products, chains, reviews, wishlists, wishlistOptions } from './data.js';

const port = Number(process.env.PORT || 3001);
const defaultLimit = 20;

const server = createServer(async (request, response) => {
    setCorsHeaders(response);

    if (request.method === 'OPTIONS') {
        response.writeHead(204);
        response.end();
        return;
    }

    const url = new URL(request.url || '/', `http://${request.headers.host || 'localhost'}`);
    try {
        if (url.pathname === '/health' && request.method === 'GET') return sendJson(response, 200, { status: 'ok' });
        if (url.pathname === '/api/v1/auth/login' && request.method === 'POST') return login(request, response);
        if (url.pathname === '/api/v1/auth/register' && request.method === 'POST') return register(request, response);

        if (url.pathname.startsWith('/api/v1/products')) return handleProducts(request, response, url);
        if (url.pathname.startsWith('/api/v1/categories')) return handleCategories(request, response, url);
        if (url.pathname.startsWith('/api/v1/customers')) return handleCustomers(request, response, url);
        if (url.pathname.startsWith('/api/v1/chains')) return handleChains(request, response, url);
        if (url.pathname.startsWith('/api/v1/reviews')) return handleReviews(request, response, url);
        if (url.pathname.startsWith('/api/v1/wishlists')) return handleWishlists(request, response, url);
        if (url.pathname === '/api/v1/search/chain' && request.method === 'GET') return findChain(response, url.searchParams);
        return sendError(response, 404, 'Ресурс не найден');
    } catch (error) {
        console.error(error);
        return sendError(response, 500, 'Внутренняя ошибка мок-API');
    }
});

async function login(request, response) {
    const body = await readJson(request);
    const customer = customers.find((item) => item.email === body?.email && item.password === body?.password && item.is_active);
    if (!customer) return sendError(response, 400, 'Неверный email или пароль');
    return sendJson(response, 200, { token: `mock-token:${customer.customer_id}` });
}

async function register(request, response) {
    const body = await readJson(request);
    if (!body?.email || !body?.password) return sendError(response, 400, 'Некорректное тело запроса');
    if (customers.some((item) => item.email === body.email)) return sendError(response, 409, 'Пользователь уже существует');
    const customer = makeCustomer(body.email, body.password);
    customers.push(customer);
    return sendJson(response, 201, publicCustomer(customer));
}

async function handleCustomers(request, response, url) {
    const parts = getResourceParts(url.pathname, '/api/v1/customers');
    if (request.method === 'POST' && parts.length === 0) return register(request, response);
    if (request.method === 'GET' && parts.length === 0) return sendJson(response, 200, activeCustomers().slice(...sliceBounds(url.searchParams)) .map(publicCustomer));
    if (parts.length !== 1) return sendError(response, 404, 'Пользователь не найден');
    const index = customers.findIndex(({ customer_id: id }) => id === parts[0]);
    if (index < 0 || !customers[index].is_active) return sendError(response, 404, 'Пользователь не найден');
    if (request.method === 'GET') return sendJson(response, 200, publicCustomer(customers[index]));
    if (request.method === 'PATCH') {
        const body = await readJson(request);
        if (!body || typeof body !== 'object') return sendError(response, 400, 'Некорректное тело запроса');
        if (body.email !== undefined) customers[index].email = body.email;
        if (body.password !== undefined) customers[index].password = body.password;
        customers[index].updated_at = new Date().toISOString();
        return sendJson(response, 200, publicCustomer(customers[index]));
    }
    if (request.method === 'DELETE') {
        customers[index].is_active = false;
        customers[index].updated_at = new Date().toISOString();
        response.writeHead(204); return response.end();
    }
    return sendError(response, 405, 'Метод не поддерживается');
}

async function handleProducts(request, response, url) {
    const parts = getResourceParts(url.pathname, '/api/v1/products');
    if (request.method === 'POST' && parts.length === 0) {
        const body = await readJson(request);
        if (!body?.customer_id || !body?.name) return sendError(response, 400, 'Некорректное тело запроса');
        const product = makeProduct(body); products.unshift(product);
        return sendJson(response, 201, product);
    }
    if (request.method === 'GET' && parts.length === 0) {
        const categoryId = url.searchParams.get('category_id');
        const list = activeProducts().filter((item) => !categoryId || item.category_id === categoryId);
        return sendJson(response, 200, list.slice(...sliceBounds(url.searchParams)));
    }
    if (request.method === 'GET' && parts[0] === 'search' && parts.length === 1) return sendProductSearch(response, url.searchParams);
    if (request.method === 'GET' && parts[0] === 'by-customer' && parts.length === 2) return sendJson(response, 200, activeProducts().filter((item) => item.customer_id === parts[1]));
    if (parts.length !== 1) return sendError(response, 404, 'Товар не найден');
    const index = products.findIndex(({ product_id: id }) => id === parts[0]);
    if (index < 0 || !products[index].is_active) return sendError(response, 404, 'Товар не найден');
    if (request.method === 'GET') return sendJson(response, 200, products[index]);
    if (request.method === 'PATCH') {
        const body = await readJson(request);
        if (!body || typeof body !== 'object') return sendError(response, 400, 'Некорректное тело запроса');
        products[index] = { ...products[index], ...pick(body, ['name', 'description', 'category_id', 'is_active']), updated_at: new Date().toISOString() };
        return sendJson(response, 200, products[index]);
    }
    if (request.method === 'DELETE') {
        products[index].is_active = false; products[index].updated_at = new Date().toISOString();
        response.writeHead(204); return response.end();
    }
    return sendError(response, 405, 'Метод не поддерживается');
}

function sendProductSearch(response, params) {
    const query = (params.get('q') || '').trim().toLocaleLowerCase('ru-RU');
    const categoryId = (params.get('category_id') || '').trim();
    if (!query && !categoryId) return sendError(response, 400, 'Некорректный поисковый запрос');
    return sendJson(response, 200, activeProducts().filter((product) => {
        const textMatch = !query || [product.name, product.description].some((value) => value.toLocaleLowerCase('ru-RU').includes(query));
        return textMatch && (!categoryId || product.category_id === categoryId);
    }));
}

async function handleCategories(request, response, url) {
    const parts = getResourceParts(url.pathname, '/api/v1/categories');
    if (request.method === 'GET' && parts.length === 0) return sendJson(response, 200, categories);
    if (request.method === 'POST' && parts.length === 0) {
        const body = await readJson(request);
        if (!body?.name) return sendError(response, 400, 'Некорректное тело запроса');
        const category = makeCategory(body); categories.push(category); return sendJson(response, 201, category);
    }
    if (request.method === 'GET' && parts.length === 2 && parts[1] === 'subcategories') return sendJson(response, 200, categories.filter((item) => item.parent_id === parts[0]));
    if (parts.length !== 1) return sendError(response, 404, 'Категория не найдена');
    const index = categories.findIndex(({ category_id: id }) => id === parts[0]);
    if (index < 0) return sendError(response, 404, 'Категория не найдена');
    if (request.method === 'GET') return sendJson(response, 200, categories[index]);
    if (request.method === 'PUT') {
        const body = await readJson(request);
        if (!body?.name) return sendError(response, 400, 'Некорректное тело запроса');
        categories[index] = { ...categories[index], name: body.name, ...(body.parent_id !== undefined ? { parent_id: body.parent_id } : {}), updated_at: new Date().toISOString() };
        return sendJson(response, 200, categories[index]);
    }
    if (request.method === 'DELETE') { categories.splice(index, 1); response.writeHead(204); return response.end(); }
    return sendError(response, 405, 'Метод не поддерживается');
}

async function handleChains(request, response, url) {
    const parts = getResourceParts(url.pathname, '/api/v1/chains');
    if (request.method === 'POST' && parts.length === 0) {
        const body = await readJson(request);
        if (!body?.from_product_id || !body?.to_product_id) return sendError(response, 400, 'Некорректное тело запроса');
        const chain = makeChain(body); chains.push(chain); return sendJson(response, 201, chain);
    }
    if (request.method === 'GET' && parts[0] === 'by-product' && parts.length === 2) return sendJson(response, 200, chains.filter((item) => item.from_product_id === parts[1] || item.to_product_id === parts[1]));
    if (parts.length !== 1 && !(parts.length === 2 && parts[1] === 'full') && !(parts.length === 2 && parts[1] === 'status')) return sendError(response, 404, 'Цепочка не найдена');
    const index = chains.findIndex(({ chain_id: id }) => id === parts[0]);
    if (index < 0) return sendError(response, 404, 'Цепочка не найдена');
    if (request.method === 'GET' && parts.length === 1) return sendJson(response, 200, chains[index]);
    if (request.method === 'GET' && parts[1] === 'full') return sendJson(response, 200, fullChain(chains[index]));
    if (request.method === 'PATCH' && parts[1] === 'status') {
        const body = await readJson(request);
        if (!body?.status) return sendError(response, 400, 'Некорректное тело запроса');
        chains[index].status = body.status; chains[index].updated_at = new Date().toISOString(); response.writeHead(204); return response.end();
    }
    if (request.method === 'DELETE') { chains.splice(index, 1); response.writeHead(204); return response.end(); }
    return sendError(response, 405, 'Метод не поддерживается');
}

async function handleReviews(request, response, url) {
    const parts = getResourceParts(url.pathname, '/api/v1/reviews');
    if (request.method === 'POST' && parts.length === 0) {
        const body = await readJson(request);
        if (!body?.from_customer_id || !body?.to_customer_id || !body?.rating) return sendError(response, 400, 'Некорректное тело запроса');
        const review = { review_id: randomUUID(), ...body, created_at: new Date().toISOString(), updated_at: new Date().toISOString() }; reviews.push(review); return sendJson(response, 201, review);
    }
    if (parts[0] === 'by-customer' && parts.length >= 2) {
        const customerReviews = reviews.filter((item) => item.to_customer_id === parts[1]);
        if (request.method === 'GET' && parts[2] === 'rating') return sendJson(response, 200, { average_rating: average(customerReviews.map((item) => item.rating)) });
        if (request.method === 'GET') return sendJson(response, 200, customerReviews);
    }
    if (parts.length !== 1) return sendError(response, 404, 'Отзыв не найден');
    const index = reviews.findIndex(({ review_id: id }) => id === parts[0]);
    if (index < 0) return sendError(response, 404, 'Отзыв не найден');
    if (request.method === 'GET') return sendJson(response, 200, reviews[index]);
    if (request.method === 'DELETE') { reviews.splice(index, 1); response.writeHead(204); return response.end(); }
    return sendError(response, 405, 'Метод не поддерживается');
}

async function handleWishlists(request, response, url) {
    const parts = getResourceParts(url.pathname, '/api/v1/wishlists');
    if (request.method === 'POST' && parts.length === 0) {
        const body = await readJson(request);
        if (!body?.product_id || !body?.name) return sendError(response, 400, 'Некорректное тело запроса');
        const wishlist = { wishlist_id: randomUUID(), ...body, created_at: new Date().toISOString(), updated_at: new Date().toISOString() }; wishlists.push(wishlist); return sendJson(response, 201, wishlist);
    }
    if (parts[0] === 'by-product' && parts.length === 2 && request.method === 'GET') {
        const wishlist = wishlists.find((item) => item.product_id === parts[1]);
        return wishlist ? sendJson(response, 200, wishlist) : sendError(response, 404, 'Список желаний не найден');
    }
    if (parts.length >= 2 && parts[1] === 'options') {
        const wishlist = wishlists.find((item) => item.wishlist_id === parts[0]);
        if (!wishlist) return sendError(response, 404, 'Список желаний не найден');
        if (request.method === 'GET' && parts.length === 2) return sendJson(response, 200, categories.filter((item) => (wishlistOptions[wishlist.wishlist_id] || []).includes(item.category_id)));
        if (request.method === 'POST' && parts.length === 2) {
            const body = await readJson(request); if (!body?.category_id) return sendError(response, 400, 'Некорректное тело запроса');
            wishlistOptions[wishlist.wishlist_id] ||= []; if (!wishlistOptions[wishlist.wishlist_id].includes(body.category_id)) wishlistOptions[wishlist.wishlist_id].push(body.category_id);
            response.writeHead(204); return response.end();
        }
        if (request.method === 'DELETE' && parts.length === 3) { wishlistOptions[wishlist.wishlist_id] = (wishlistOptions[wishlist.wishlist_id] || []).filter((id) => id !== parts[2]); response.writeHead(204); return response.end(); }
    }
    if (parts.length !== 1) return sendError(response, 404, 'Список желаний не найден');
    const index = wishlists.findIndex(({ wishlist_id: id }) => id === parts[0]);
    if (index < 0) return sendError(response, 404, 'Список желаний не найден');
    if (request.method === 'GET') return sendJson(response, 200, wishlists[index]);
    if (request.method === 'DELETE') { delete wishlistOptions[wishlists[index].wishlist_id]; wishlists.splice(index, 1); response.writeHead(204); return response.end(); }
    return sendError(response, 405, 'Метод не поддерживается');
}

function findChain(response, params) {
    const target = params.get('target_product_id');
    if (!target) return sendError(response, 400, 'Некорректный поисковый запрос');
    const result = chains.find((item) => item.to_product_id === target && item.status !== 'cancelled');
    const chain = result ? [products.find((item) => item.product_id === result.from_product_id), products.find((item) => item.product_id === result.to_product_id)].filter(Boolean) : [];
    return sendJson(response, 200, { chain, length: chain.length });
}

function fullChain(chain) { return chains.filter((item) => item.chain_id === chain.chain_id || item.chain_id === chain.previous_chain_id || item.chain_id === chain.next_chain_id); }
function makeCustomer(email, password) { const now = new Date().toISOString(); return { customer_id: randomUUID(), email, password, is_active: true, created_at: now, updated_at: now }; }
function makeProduct(body) { const now = new Date().toISOString(); return { product_id: randomUUID(), customer_id: body.customer_id, ...(body.category_id ? { category_id: body.category_id } : {}), name: body.name, ...(body.description ? { description: body.description } : {}), is_active: true, created_at: now, updated_at: now }; }
function makeCategory(body) { const now = new Date().toISOString(); return { category_id: randomUUID(), name: body.name, ...(body.parent_id ? { parent_id: body.parent_id } : {}), created_at: now, updated_at: now }; }
function makeChain(body) { const now = new Date().toISOString(); return { chain_id: randomUUID(), from_product_id: body.from_product_id, to_product_id: body.to_product_id, initiator_id: body.initiator_id || 'user-pskov-01', ...(body.message ? { message: body.message } : {}), status: body.status || 'pending', created_at: now, updated_at: now }; }
function publicCustomer(customer) {
    const result = { ...customer };
    delete result.password;
    delete result.is_active;
    return result;
}
function activeCustomers() { return customers.filter((item) => item.is_active); }
function activeProducts() { return products.filter((item) => item.is_active); }
function average(values) { return values.length ? values.reduce((sum, value) => sum + value, 0) / values.length : 0; }
function pick(value, keys) { return Object.fromEntries(keys.filter((key) => value[key] !== undefined).map((key) => [key, value[key]])); }
function getResourceParts(pathname, prefix) { return pathname.slice(prefix.length).split('/').filter(Boolean); }
function sliceBounds(params) { const offset = Math.max(0, Number.parseInt(params.get('offset') || '0', 10) || 0); const limit = Math.min(100, Math.max(1, Number.parseInt(params.get('limit') || String(defaultLimit), 10) || defaultLimit)); return [offset, offset + limit]; }
function readJson(request) { return new Promise((resolve) => { let body = ''; request.setEncoding('utf8'); request.on('data', (chunk) => { body += chunk; }); request.on('end', () => { try { resolve(body ? JSON.parse(body) : undefined); } catch { resolve(undefined); } }); }); }
function setCorsHeaders(response) { response.setHeader('Access-Control-Allow-Origin', '*'); response.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization'); response.setHeader('Access-Control-Allow-Methods', 'DELETE, GET, OPTIONS, PATCH, POST, PUT'); }
function sendJson(response, status, body) { response.writeHead(status, { 'Content-Type': 'application/json; charset=utf-8' }); response.end(JSON.stringify(body)); }
function sendError(response, status, message) { sendJson(response, status, { error: message }); }

server.listen(port, () => console.log(`Mock API слушает http://localhost:${port}`));

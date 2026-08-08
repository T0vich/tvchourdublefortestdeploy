export type TProductListRequest = {
    category_id?: string;
    offset?: number;
    limit?: number;
};

/**
 * Статусы товара из docs/API.md: товар свободен, зарезервирован под сделку,
 * уже обменян или снят с публикации.
 */
export type TProductStatus = 'active' | 'reserved' | 'exchanged' | 'archived';

export type TCreateProductRequest = {
    customer_id: string;
    category_id?: string;
    title: string;
    description?: string;
    image?: string;
    price?: number;
    location?: string;
};

export type TUpdateProductRequest = {
    category_id?: string;
    title?: string;
    description?: string;
    image?: string;
    price?: number;
    location?: string;
    status?: TProductStatus;
};

export type TProduct = {
    product_id: string;
    customer_id: string;
    category_id?: string;
    title: string;
    description?: string;
    image?: string;
    price?: number;
    location?: string;
    status: TProductStatus;
    created_at: string;
    updated_at: string;
};

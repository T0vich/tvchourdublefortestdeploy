import Styles from './catalog-page.module.css';
import {MainSection} from '@shared/ui/mainSection';
import {ProductCard} from '@shared/ui/productCard';
import {useGetProductsQuery} from '@entities/product';
import {usePageTitle} from "@app/providers/pageTitle";
import {useLayoutEffect} from "react";
import {Preloader} from "@shared/ui/preloader";

export const CatalogPage = () => {
    const {setTitle} = usePageTitle();

    const {
        data,
        isLoading,
        isError,
    } = useGetProductsQuery({
        offset: 0,
        limit: 20,
    });

    useLayoutEffect(() => {
        setTitle('Вещи в обороте');
    }, [setTitle]);

    if (isLoading) {
        return <Preloader/>;
    }

    if (isError) {
        return <p>Не удалось загрузить товары</p>;
    }

    return (
        <MainSection>
            <div className={Styles['catalog-page']}>
                {data?.map((product) => (
                    <ProductCard
                        key={product.product_id}
                        title={product.title}
                        img={product.image}
                        price={product.price}
                        location={product.location}
                    />
                ))}
            </div>
        </MainSection>
    );
};

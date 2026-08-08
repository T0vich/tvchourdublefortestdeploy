import { useContext } from 'react';

import { PageTitleContext } from './pageTitleContext';

export const usePageTitle = () => {
    const context = useContext(PageTitleContext);

    if (!context) {
        throw new Error('usePageTitle должен использоваться внутри PageTitleProvider');
    }

    return context;
};

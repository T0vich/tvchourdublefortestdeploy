import { useState, type PropsWithChildren } from 'react';

import { PageTitleContext } from './pageTitleContext';

export const PageTitleProvider = ({children}: PropsWithChildren) => {
    const [title, setTitle] = useState('');

    return (
        <PageTitleContext.Provider value={{title, setTitle}}>
            {children}
        </PageTitleContext.Provider>
    );
};

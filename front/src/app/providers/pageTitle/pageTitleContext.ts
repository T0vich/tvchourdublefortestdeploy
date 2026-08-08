import { createContext } from 'react';

export type PageTitleContextValue = {
    title: string;
    setTitle: (title: string) => void;
};

export const PageTitleContext = createContext<PageTitleContextValue | null>(null);

import type { TUserProfile } from '../types';

export const getUserLocalStorage = () => {
    const userFromStorage = localStorage.getItem('user');

    if (!userFromStorage) {
        return undefined;
    }

    try {
        return JSON.parse(userFromStorage) as TUserProfile;
    } catch {
        localStorage.removeItem('user');
        return undefined;
    }
};

export const setUserLocalStorage = (userData: TUserProfile) => {
    try {
        localStorage.setItem('user', JSON.stringify(userData));
    } catch {
        return;
    }
};

export const removeUserLocalStorage = () => {
    localStorage.removeItem('user');
    localStorage.removeItem('token');
};

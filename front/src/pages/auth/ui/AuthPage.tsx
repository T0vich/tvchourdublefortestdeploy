import { useLayoutEffect } from 'react';

import { AuthForm } from '@features/auth';
import { usePageTitle } from '@app/providers/pageTitle';

import Styles from './auth-page.module.css';

export const AuthPage = () => {
    const {setTitle} = usePageTitle();

    useLayoutEffect(() => {
        setTitle('Вход и регистрация');
    }, [setTitle]);

    return (
        <section className={Styles.page}>
            <div className={Styles.card}>
                <h1>Добро пожаловать</h1>
                <p className={Styles.description}>
                    Войдите, чтобы обмениваться вещами и находить нужные товары.
                </p>
                <AuthForm />
            </div>
        </section>
    );
};

import { FormEvent, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import {
    useLoginUserMutation,
    useRegisterUserMutation,
} from '@entities/user';
import { Button } from '@shared/ui/button';
import { Input } from '@shared/ui/input';

import Styles from './auth-form.module.css';

type TMode = 'login' | 'register';
type TField = 'email' | 'password' | 'confirmPassword';
type TErrors = Partial<Record<TField, string>>;

const getErrorMessage = (error: unknown) => {
    if (typeof error === 'object' && error !== null && 'data' in error) {
        const data = error.data;
        if (typeof data === 'object' && data !== null && 'error' in data && typeof data.error === 'string') {
            // При ответе 422 сервер присылает разбор по полям — показываем его,
            // иначе пользователь видит только «данные не прошли проверку».
            if ('fields' in data && typeof data.fields === 'object' && data.fields !== null) {
                const details = Object.values(data.fields as Record<string, string>).join('; ');
                if (details) {
                    return `${data.error}: ${details}`;
                }
            }
            return data.error;
        }
    }
    return 'Не удалось выполнить запрос. Попробуйте ещё раз.';
};

const validate = (email: string, password: string, confirmPassword: string, mode: TMode): TErrors => {
    const errors: TErrors = {};

    if (!email.trim()) {
        errors.email = 'Введите email';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
        errors.email = 'Введите корректный email';
    }

    if (!password) {
        errors.password = 'Введите пароль';
    } else if (password.length < 8) {
        errors.password = 'Пароль должен содержать минимум 8 символов';
    }

    if (mode === 'register' && password !== confirmPassword) {
        errors.confirmPassword = 'Пароли не совпадают';
    }

    return errors;
};

export const AuthForm = () => {
    const navigate = useNavigate();
    const [mode, setMode] = useState<TMode>('login');
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [errors, setErrors] = useState<TErrors>({});
    const [requestError, setRequestError] = useState<string>();
    const [successMessage, setSuccessMessage] = useState<string>();
    const [loginUser, { isLoading: isLoginLoading }] = useLoginUserMutation();
    const [registerUser, { isLoading: isRegisterLoading }] = useRegisterUserMutation();
    const isLoading = isLoginLoading || isRegisterLoading;

    const switchMode = () => {
        setMode((currentMode) => currentMode === 'login' ? 'register' : 'login');
        setErrors({});
        setRequestError(undefined);
        setSuccessMessage(undefined);
        setConfirmPassword('');
    };

    const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
        event.preventDefault();
        setRequestError(undefined);
        setSuccessMessage(undefined);

        const validationErrors = validate(email, password, confirmPassword, mode);
        setErrors(validationErrors);
        if (Object.keys(validationErrors).length > 0) {
            return;
        }

        try {
            if (mode === 'login') {
                const response = await loginUser({email: email.trim(), password}).unwrap();
                localStorage.setItem('token', response.token);
                navigate('/');
                return;
            }

            // Регистрация возвращает токен сразу, поэтому просить пользователя
            // тут же войти теми же данными не нужно.
            const response = await registerUser({email: email.trim(), password}).unwrap();
            localStorage.setItem('token', response.token);
            navigate('/');
        } catch (error) {
            setRequestError(getErrorMessage(error));
        }
    };

    return (
        <form className={Styles.form} onSubmit={handleSubmit} noValidate>
            <div className={Styles.fields}>
                <Input
                    label="Email"
                    name="email"
                    type="email"
                    value={email}
                    placeholder="you@example.com"
                    onChange={setEmail}
                    disabled={isLoading}
                    error={{showError: Boolean(errors.email), errorMessage: errors.email ?? ''}}
                />
                <Input
                    label="Пароль"
                    name="password"
                    type="password"
                    value={password}
                    placeholder="Минимум 8 символов"
                    onChange={setPassword}
                    disabled={isLoading}
                    error={{showError: Boolean(errors.password), errorMessage: errors.password ?? ''}}
                />
                {mode === 'register' && (
                    <Input
                        label="Повторите пароль"
                        name="confirmPassword"
                        type="password"
                        value={confirmPassword}
                        placeholder="Повторите пароль"
                        onChange={setConfirmPassword}
                        disabled={isLoading}
                        error={{showError: Boolean(errors.confirmPassword), errorMessage: errors.confirmPassword ?? ''}}
                    />
                )}
            </div>

            {requestError && <p className={Styles.error}>{requestError}</p>}
            {successMessage && <p className={Styles.success}>{successMessage}</p>}

            <Button type="submit" loading={isLoading}>
                {mode === 'login' ? 'Войти' : 'Зарегистрироваться'}
            </Button>

            <Button type="button" variant="text" onClick={switchMode} disabled={isLoading}>
                {mode === 'login' ? 'Создать аккаунт' : 'Уже есть аккаунт? Войти'}
            </Button>
        </form>
    );
};

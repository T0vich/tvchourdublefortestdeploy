import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type {
    TLoginPayload,
    TRegisterPayload,
    TAuthResponse,
    TUser,
} from '../types';
import { getApiBaseUrl } from '@/shared/api';

export const userApi = createApi({
    reducerPath: 'userApi',
    baseQuery: fetchBaseQuery({ baseUrl: `${getApiBaseUrl()}/api/v1` }),
    endpoints: (builder) => ({
        loginUser: builder.mutation<TAuthResponse, TLoginPayload>({
            query: (body) => ({url: '/auth/login', method: 'POST', body}),
        }),
        registerUser: builder.mutation<TAuthResponse, TRegisterPayload>({
            query: (body) => ({url: '/auth/register', method: 'POST', body}),
        }),
        // Профиль владельца токена: заодно проверка, что сохранённый токен ещё жив.
        getMe: builder.query<TUser, void>({
            query: () => ({
                url: '/auth/me',
                headers: {Authorization: `Bearer ${localStorage.getItem('token') ?? ''}`},
            }),
        }),
    }),
});

export const {
    useLoginUserMutation,
    useRegisterUserMutation,
    useGetMeQuery,
} = userApi;

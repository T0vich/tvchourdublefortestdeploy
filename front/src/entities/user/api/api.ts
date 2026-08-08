import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type {
    TLoginPayload,
    TRegisterPayload,
    TAuthResponse,
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
    }),
});

export const {
    useLoginUserMutation,
    useRegisterUserMutation,
} = userApi;

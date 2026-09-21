import client from './client'; import type { ApiResponse, Greenhouse } from '../types/domain';
export const listGreenhouses = async () => (await client.get<ApiResponse<Greenhouse[]>>('/greenhouses')).data.data;
export const getGreenhouse = async (id: number) => (await client.get<ApiResponse<Greenhouse>>(`/greenhouses/${id}`)).data.data;
export const simulate = async (id: number) => (await client.post(`/greenhouses/${id}/simulate`)).data.data;

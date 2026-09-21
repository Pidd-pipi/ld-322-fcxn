import client from './client'; import type { Alert, ApiResponse, Device, EnvironmentReport, Reading } from '../types/domain';
export const getLatest = async (id:number) => (await client.get<ApiResponse<Reading[]>>('/readings/latest',{params:{greenhouse_id:id}})).data.data;
export const getHistory = async (id:number,range:string,types?:string[]) => { const end=new Date(); const start=new Date(end.getTime()-(range==='week'?7:range==='month'?30:1)*86400000); return (await client.get<ApiResponse<Reading[]>>('/readings/history',{params:{greenhouse_id:id,start:start.toISOString(),end:end.toISOString(),types:types?.join(',')}})).data.data; };
export const getAlerts = async (id?:number) => (await client.get<ApiResponse<Alert[]>>('/alerts',{params:id?{greenhouse_id:id}:{}})).data.data;
export const handleAlert = async (id:number) => (await client.patch(`/alerts/${id}/handle`)).data.data;
export const getDevices = async (id:number) => (await client.get<ApiResponse<Device[]>>('/devices',{params:{greenhouse_id:id}})).data.data;
export const toggleDevice = async (id:number,status:'on'|'off') => (await client.patch<ApiResponse<Device>>(`/devices/${id}/toggle`,{status})).data.data;
export const getReport = async (id:number,range:string) => (await client.get<ApiResponse<EnvironmentReport>>('/reports/environment',{params:{greenhouse_id:id,range}})).data.data;
export const login = async () => (await client.post<ApiResponse<{token:string}>>('/auth/login',{username:'admin',password:'admin123'})).data.data;

import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
export default defineConfig({ plugins: [react()], server: { port: 18622, host: '0.0.0.0', proxy: { '/api': 'http://localhost:19622', '/ws': { target: 'ws://localhost:19622', ws: true } } } });

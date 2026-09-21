import { useEffect } from 'react';
export function useWebSocket(onEvent: () => void) { useEffect(() => { const protocol=location.protocol==='https:'?'wss':'ws'; const ws=new WebSocket(`${protocol}://${location.host}/ws`); ws.onmessage=onEvent; return () => ws.close(); }, [onEvent]); }

const LOCAL_API_ORIGIN = 'http://localhost:8080';
const LOCAL_WS_URL = 'ws://localhost:8080/ws';

const isLocalhost = () => {
  if (typeof window === 'undefined') {
    return false;
  }
  return window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
};

export const API_BASE_URL = isLocalhost() ? LOCAL_API_ORIGIN : '';

export const API_URL = `${API_BASE_URL}/api`;

const wsProtocol = typeof window !== 'undefined' && window.location.protocol === 'https:' ? 'wss' : 'ws';
const wsHost = typeof window !== 'undefined' ? window.location.host : 'localhost:8080';

export const WS_URL = isLocalhost()
  ? LOCAL_WS_URL
  : `${wsProtocol}://${wsHost}/ws`;

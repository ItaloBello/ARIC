import AsyncStorage from '@react-native-async-storage/async-storage';

const TOKEN_KEY = '@aric/api/jwt/v1';
const configuredBaseUrl = (process.env.EXPO_PUBLIC_API_URL ?? '').trim().replace(/\/+$/, '');

export const apiConfigured = Boolean(configuredBaseUrl);
export const apiBaseUrl = configuredBaseUrl;

export type ApiError = Error & { status?: number; retryAfter?: string };
let unauthorizedHandler: (() => void) | null = null;

export function setUnauthorizedHandler(handler: (() => void) | null) {
  unauthorizedHandler = handler;
}

export async function getApiToken() {
  return AsyncStorage.getItem(TOKEN_KEY);
}

export async function setApiToken(token: string | null) {
  if (token) await AsyncStorage.setItem(TOKEN_KEY, token);
  else await AsyncStorage.removeItem(TOKEN_KEY);
}

export async function apiRequest<T>(path: string, options: RequestInit = {}, authenticated = true): Promise<T> {
  if (!apiConfigured) {
    throw new Error('A API ainda não foi configurada. Defina EXPO_PUBLIC_API_URL no arquivo .env.');
  }

  const token = authenticated ? await getApiToken() : null;
  const headers = new Headers(options.headers ?? {});
  if (options.body && !(options.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }
  if (token) headers.set('Authorization', `Bearer ${token}`);

  let response: Response;
  try {
    response = await fetch(`${configuredBaseUrl}${path.startsWith('/') ? path : `/${path}`}`, {
      ...options,
      headers,
    });
  } catch {
    throw new Error('Não foi possível alcançar o servidor ARIC. Verifique a URL da API e a conexão de rede.');
  }

  if (response.status === 204) return undefined as T;

  const contentType = response.headers.get('content-type') ?? '';
  const payload = contentType.includes('application/json')
    ? await response.json().catch(() => null)
    : await response.text().catch(() => '');

  if (!response.ok) {
    if (response.status === 401 && authenticated && token) {
      await setApiToken(null);
      unauthorizedHandler?.();
    }
    const message = payload && typeof payload === 'object' && 'error' in payload
      ? String((payload as { error: unknown }).error)
      : typeof payload === 'string' && payload.trim()
        ? payload
        : `Erro ${response.status} ao comunicar com a API.`;
    const error = new Error(message) as ApiError;
    error.status = response.status;
    error.retryAfter = response.headers.get('Retry-After') ?? undefined;
    throw error;
  }

  return payload as T;
}

export function apiMediaUrl(path?: string | null) {
  if (!path) return undefined;
  if (/^https?:\/\//i.test(path)) return path;
  if (!apiConfigured) return undefined;
  const origin = configuredBaseUrl.replace(/\/api\/v1\/?$/, '');
  return `${origin}${path.startsWith('/') ? path : `/${path}`}`;
}

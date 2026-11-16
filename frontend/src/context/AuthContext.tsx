import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { apiConfigured, apiRequest, getApiToken, setApiToken, setUnauthorizedHandler } from '../services/api';

type AuthStatus = 'loading' | 'signedOut' | 'signedIn' | 'guest';
type Provider = 'api' | 'guest';

type ApiUser = {
  id: number;
  nome: string;
  email: string;
  fotoUrl?: string | null;
  bio?: string | null;
  bairro?: string | null;
  cidade?: string | null;
  uf?: string | null;
  papel?: string | null;
  xpTotal?: number;
  sequenciaDias?: number;
};

type Session = {
  uid: string;
  email?: string;
  displayName?: string;
  provider: Provider;
  isAdmin: boolean;
  emailVerified: boolean;
  papel?: string;
};

type AuthResponse = { token: string; expiraEm: string; usuario: ApiUser };
type AuthResult = { ok: true } | { ok: false; message: string };

type AuthContextValue = {
  status: AuthStatus;
  session: Session | null;
  apiConfigured: boolean;
  isGuest: boolean;
  isAdmin: boolean;
  uid: string | null;
  emailVerified: boolean;
  login: (email: string, password: string) => Promise<AuthResult>;
  register: (name: string, email: string, password: string, cep: string, numeroEndereco: string, bairro?: string, logradouro?: string, complementoEndereco?: string) => Promise<AuthResult>;
  resetPassword: (email: string) => Promise<AuthResult>;
  continueAsGuest: () => Promise<void>;
  signOut: () => Promise<void>;
  deleteAccount: (password: string, apagarConteudo: boolean) => Promise<AuthResult>;
  refreshRole: () => Promise<void>;
  sendVerificationEmail: () => Promise<AuthResult>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

function sessionFromUser(user: ApiUser): Session {
  const role = String(user.papel ?? '').toLowerCase();
  return {
    uid: String(user.id),
    email: user.email,
    displayName: user.nome,
    provider: 'api',
    isAdmin: role === 'admin' || role === 'moderador',
    emailVerified: true,
    papel: role || 'usuario',
  };
}

function friendlyApiError(error: unknown) {
  const message = error instanceof Error ? error.message : 'Não foi possível concluir a operação.';
  if (message.includes('muitas tentativas')) return message;
  if (message.toLowerCase().includes('e-mail já cadastrado')) return 'Este e-mail já possui uma conta.';
  if (message.toLowerCase().includes('senha')) return message;
  return message;
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [status, setStatus] = useState<AuthStatus>('loading');
  const [session, setSession] = useState<Session | null>(null);

  useEffect(() => {
    let active = true;
    setUnauthorizedHandler(() => { setSession(null); setStatus('signedOut'); });
    (async () => {
      try {
        const token = await getApiToken();
        if (!token) {
          if (active) setStatus('signedOut');
          return;
        }
        const user = await apiRequest<ApiUser>('/auth/me');
        if (active) {
          setSession(sessionFromUser(user));
          setStatus('signedIn');
        }
      } catch {
        await setApiToken(null);
        if (active) {
          setSession(null);
          setStatus('signedOut');
        }
      }
    })();
    return () => { active = false; setUnauthorizedHandler(null); };
  }, []);

  const login = useCallback(async (email: string, password: string): Promise<AuthResult> => {
    const normalizedEmail = email.trim().toLowerCase();
    if (!/^\S+@\S+\.\S+$/.test(normalizedEmail)) return { ok: false, message: 'Digite um e-mail válido.' };
    if (!password) return { ok: false, message: 'Informe sua senha.' };
    try {
      const result = await apiRequest<AuthResponse>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email: normalizedEmail, senha: password }),
      }, false);
      await setApiToken(result.token);
      setSession(sessionFromUser(result.usuario));
      setStatus('signedIn');
      return { ok: true };
    } catch (error) {
      return { ok: false, message: friendlyApiError(error) };
    }
  }, []);

  const register = useCallback(async (name: string, email: string, password: string, cep: string, numeroEndereco: string, bairro = '', logradouro = '', complementoEndereco = ''): Promise<AuthResult> => {
    const cleanName = name.trim();
    const normalizedEmail = email.trim().toLowerCase();
    const cleanCep = cep.replace(/\D/g, '');
    if (cleanName.length < 2) return { ok: false, message: 'Informe seu nome.' };
    if (!/^\S+@\S+\.\S+$/.test(normalizedEmail)) return { ok: false, message: 'Digite um e-mail válido.' };
    if (!/^[0-9]{8}$/.test(cleanCep)) return { ok: false, message: 'Informe um CEP válido com 8 números.' };
    if (!numeroEndereco.trim()) return { ok: false, message: 'Informe o número do endereço.' };
    if (password.length < 8 || password.length > 72 || !/[A-Z]/.test(password) || !/[a-z]/.test(password) || !/[0-9]/.test(password) || !/[^A-Za-z0-9]/.test(password)) {
      return { ok: false, message: 'A senha deve ter 8 a 72 caracteres, com maiúscula, minúscula, número e caractere especial.' };
    }
    try {
      const result = await apiRequest<AuthResponse>('/auth/registro', {
        method: 'POST',
        body: JSON.stringify({ nome: cleanName, email: normalizedEmail, senha: password, confirmarSenha: password, cep: cleanCep, numeroEndereco: numeroEndereco.trim(), bairro: bairro.trim(), logradouro: logradouro.trim(), complementoEndereco: complementoEndereco.trim() }),
      }, false);
      await setApiToken(result.token);
      setSession(sessionFromUser(result.usuario));
      setStatus('signedIn');
      return { ok: true };
    } catch (error) {
      return { ok: false, message: friendlyApiError(error) };
    }
  }, []);

  const resetPassword = useCallback(async (_email: string): Promise<AuthResult> => ({
    ok: false,
    message: 'A API Go ainda não possui uma rota de recuperação de senha. Esse recurso precisa ser implementado no backend.',
  }), []);

  const sendVerificationEmail = useCallback(async (): Promise<AuthResult> => ({
    ok: false,
    message: 'A API Go não possui verificação de e-mail. A autenticação atual não exige essa etapa.',
  }), []);

  const continueAsGuest = useCallback(async () => {
    setSession({ uid: 'guest', displayName: 'Visitante', provider: 'guest', isAdmin: false, emailVerified: false });
    setStatus('guest');
  }, []);

  const signOut = useCallback(async () => {
    await setApiToken(null);
    setSession(null);
    setStatus('signedOut');
  }, []);

  const deleteAccount = useCallback(async (password: string, apagarConteudo: boolean): Promise<AuthResult> => {
    if (!password) return { ok: false, message: 'Informe sua senha para confirmar a exclusão.' };
    try {
      await apiRequest<void>('/auth/me', {
        method: 'DELETE',
        body: JSON.stringify({ senha: password, apagarConteudo }),
      });
      await setApiToken(null);
      setSession(null);
      setStatus('signedOut');
      return { ok: true };
    } catch (error) {
      return { ok: false, message: friendlyApiError(error) };
    }
  }, []);

  const refreshRole = useCallback(async () => {
    try {
      const user = await apiRequest<ApiUser>('/auth/me');
      setSession(sessionFromUser(user));
    } catch {
      await setApiToken(null);
      setSession(null);
      setStatus('signedOut');
    }
  }, []);

  const value = useMemo<AuthContextValue>(() => ({
    status,
    session,
    apiConfigured,
    isGuest: status === 'guest',
    isAdmin: session?.isAdmin === true,
    uid: session?.uid ?? null,
    emailVerified: session?.emailVerified === true,
    login,
    register,
    resetPassword,
    continueAsGuest,
    signOut,
    deleteAccount,
    refreshRole,
    sendVerificationEmail,
  }), [continueAsGuest, deleteAccount, login, refreshRole, register, resetPassword, sendVerificationEmail, session, signOut, status]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) throw new Error('useAuth must be used inside AuthProvider');
  return context;
}

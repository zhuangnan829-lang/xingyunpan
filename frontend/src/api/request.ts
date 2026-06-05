import axios, { AxiosError, AxiosResponse, InternalAxiosRequestConfig } from 'axios';
import { errorLogger } from '@/utils/error-logger';
import { getToken, removeToken } from '@/utils/auth';
import { resolveAPIBaseURL } from '@/utils/public-url';

type RequestConfig = Parameters<typeof axios.create>[0];

interface RequestInstance {
  get<T = any>(url: string, config?: any): Promise<T>;
  delete<T = any>(url: string, config?: any): Promise<T>;
  head<T = any>(url: string, config?: any): Promise<T>;
  options<T = any>(url: string, config?: any): Promise<T>;
  post<T = any>(url: string, data?: any, config?: any): Promise<T>;
  put<T = any>(url: string, data?: any, config?: any): Promise<T>;
  patch<T = any>(url: string, data?: any, config?: any): Promise<T>;
  interceptors: ReturnType<typeof axios.create>['interceptors'];
}

function handleAuthExpired(message?: string): never {
  removeToken();

  errorLogger.warn('Authentication failed - token expired or invalid', {
    status: 401,
  });

  if (window.location.pathname !== '/login') {
    window.location.href = '/login';
  }

  throw new Error(message || '登录已过期，请重新登录。');
}

const rawRequest = axios.create({
  baseURL: resolveAPIBaseURL(import.meta.env.VITE_API_BASE_URL),
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
} satisfies RequestConfig);

function isBackendUnavailableError(error: AxiosError): boolean {
  if (error.response) {
    return false;
  }

  const requestBaseURL = String(error.config?.baseURL || rawRequest.defaults.baseURL || '');
  return /localhost:8080|127\.0\.0\.1:8080/.test(requestBaseURL);
}

rawRequest.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = getToken();
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error: AxiosError) => Promise.reject(error),
);

rawRequest.interceptors.response.use(
  (response: AxiosResponse) => {
    const { data } = response;

    if (data && typeof data === 'object' && 'code' in data) {
      if (data.code >= 200 && data.code < 300) {
        return data.data;
      }

      if (data.code === 401) {
        handleAuthExpired(data.message);
      }

      return Promise.reject(new Error(data.message || '操作失败，请稍后重试。'));
    }

    return response.data;
  },
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      handleAuthExpired('登录已过期，请重新登录。');
    }

    errorLogger.error('API request failed', error, {
      url: error.config?.url,
      method: error.config?.method,
      status: error.response?.status,
      data: error.response?.data,
    });

    const errorMessage = handleAPIError(error);
    return Promise.reject(new Error(errorMessage));
  },
);

export function handleAPIError(error: unknown): string {
  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError<any>;

    if (!axiosError.response) {
      if (axiosError.code === 'ECONNABORTED') {
        return '请求超时：后端服务响应过慢或当前不可用，请稍后重试。';
      }

      if (isBackendUnavailableError(axiosError)) {
        return '后端服务未启动或 127.0.0.1:8080 无法连接，请先启动后端服务。';
      }

      return '网络连接失败，请检查网络或确认后端服务是否已启动。';
    }

    const status = axiosError.response.status;
    const responseData = axiosError.response.data;

    if (status >= 500) {
      return '后端服务发生错误，请稍后重试或检查服务日志。';
    }

    if (status >= 400 && status < 500) {
      if (responseData && typeof responseData === 'object') {
        if ('message' in responseData && responseData.message) {
          return responseData.message;
        }
        if ('error' in responseData && responseData.error) {
          return responseData.error;
        }
      }

      return '请求失败，请检查参数或稍后重试。';
    }
  }

  console.error('Unexpected error:', error);
  return '请求失败，请稍后重试。';
}

const request = rawRequest as unknown as RequestInstance;

export default request;

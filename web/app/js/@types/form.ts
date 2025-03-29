import type { APIResponse } from './api';

export type FormValues = {
  [key: string]: string | number | boolean;
};

export type FormErrors<T> = {
  [K in keyof T]?: string;
};

export type FormOptions = {
  data: FormValues;
  method?: 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  submitUrl: string;
  errors: FormErrors<FormValues>;
  validateFn(): FormErrors<FormValues>;
  onSuccess(data: APIResponse): void;
  onError(error: unknown): void;
};

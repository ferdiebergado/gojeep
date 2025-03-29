import type { APIResponse } from '../@types/api';
import type { FormOptions } from '../@types/form';

export default function (opts: FormOptions) {
  const { data, method, submitUrl, errors, validateFn, onSuccess, onError } =
    opts;

  return {
    data,
    method: method ?? 'POST',
    submitUrl,
    isSubmitting: false,
    errors,
    message: '',
    isValid: true,
    validate(): boolean {
      this.errors = validateFn.call(this);
      const isValid = Object.values(this.errors).every((value) => value === '');
      if (!isValid) {
        this.message = 'Invalid input.';
        this.isValid = false;
      }
      return isValid;
    },
    async submit(): Promise<void> {
      this.isValid = true;
      if (!this.validate()) return;

      this.isSubmitting = true;

      try {
        const response = await fetch(this.submitUrl, {
          method: this.method,
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(this.data),
        });

        if (!response.ok) {
          const { status } = response;
          if (status === 400 || status === 422) {
            const data: APIResponse = await response.json();
            const { message, errors } = data;
            if (errors) this.errors = errors;
            throw new Error(message);
          }
          throw new Error('Invalid credentials');
        }

        const data: APIResponse = await response.json();

        this.message = data.message;

        onSuccess(data);
        this.data = {};
      } catch (error) {
        console.error(error);
        this.isValid = false;
        if (error instanceof Error) this.message = error.message;
        onError(error);
      } finally {
        this.isSubmitting = false;
        console.log('haserrors', this.isValid);
      }
    },
  };
}

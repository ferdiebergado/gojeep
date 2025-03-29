import type { FormErrors } from '../@types/form';
import { isValidEmail } from '../utils';
import form from './form';
import urls from '../endpoints';

type Values = {
  email: string;
  password: string;
  password_confirm: string;
};

type Errors = FormErrors<Values>;

function validateFormValues(data: Values): Errors {
  const { email, password, password_confirm } = data;
  const formErrors: Errors = {};

  if (!email) {
    formErrors.email = 'Email is required.';
  } else if (!isValidEmail(email)) {
    formErrors.email = 'Invalid email format.';
  }

  if (!password) {
    formErrors.password = 'Password is required.';
  }

  if (!password_confirm) {
    formErrors.password_confirm = 'Password confirmation is required.';
  } else if (password && password_confirm !== password) {
    formErrors.password_confirm = 'Passwords should match.';
  }

  return formErrors;
}

export default function () {
  const data: Values = {
    email: '',
    password: '',
    password_confirm: '',
  };

  const errors: Errors = {
    email: '',
    password: '',
    password_confirm: '',
  };

  return form({
    data,
    submitUrl: urls.register,
    errors,
    validateFn() {
      return validateFormValues(this.data as Values);
    },
    onSuccess(res) {
      const { message, data } = res;
      console.log(message, data);
    },
    onError() {
      return;
    },
  });
}

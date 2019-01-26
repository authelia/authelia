import ConfirmationSentView from "../views/ConfirmationSentView/ConfirmationSentView";
import OneTimePasswordRegistrationView from "../containers/views/OneTimePasswordRegistrationView/OneTimePasswordRegistrationView";
import SecurityKeyRegistrationView from "../containers/views/SecurityKeyRegistrationView/SecurityKeyRegistrationView";
import ForgotPasswordView from "../containers/views/ForgotPasswordView/ForgotPasswordView";
import ResetPasswordView from "../containers/views/ResetPasswordView/ResetPasswordView";
import AuthenticationView from "../containers/views/AuthenticationView/AuthenticationView";

export const routes = [{
  path: '/',
  title: 'Sign in',
  component: AuthenticationView,
}, {
  path: '/confirmation-sent',
  title: 'e-mail sent',
  component: ConfirmationSentView
}, {
  path: '/one-time-password-registration',
  title: 'One-time password',
  component: OneTimePasswordRegistrationView,
}, {
  path: '/security-key-registration',
  title: 'Security key',
  component: SecurityKeyRegistrationView,
}, {
  path: '/forgot-password',
  title: 'Forgot password',
  component: ForgotPasswordView,
}, {
  path: '/reset-password',
  title: 'Reset password',
  component: ResetPasswordView,
}]
import FirstFactorView from "../containers/views/FirstFactorView/FirstFactorView";
import SecondFactorView from "../containers/views/SecondFactorView/SecondFactorView";
import ConfirmationSentView from "../views/ConfirmationSentView/ConfirmationSentView";
import OneTimePasswordRegistrationView from "../containers/views/OneTimePasswordRegistrationView/OneTimePasswordRegistrationView";
import SecurityKeyRegistrationView from "../containers/views/SecurityKeyRegistrationView/SecurityKeyRegistrationView";
import ForgotPasswordView from "../views/ForgotPasswordView/ForgotPasswordView";
import ResetPasswordView from "../views/ResetPasswordView/ResetPasswordView";

export const routes = [{
  path: '/',
  title: 'Login',
  component: FirstFactorView,
}, {
  path: '/2fa',
  title: '2-factor',
  component: SecondFactorView,
}, {
  path: '/confirmation-sent',
  title: 'e-mail sent',
  component: ConfirmationSentView
}, {
  path: '/one-time-password-registration',
  title: 'One-time password registration',
  component: OneTimePasswordRegistrationView,
}, {
  path: '/security-key-registration',
  title: 'Security key registration',
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
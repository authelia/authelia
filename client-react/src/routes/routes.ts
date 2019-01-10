import FirstFactorView from "../views/FirstFactorView/FirstFactorView";
import SecondFactorView from "../views/SecondFactorView/SecondFactorView";
import ConfirmationSent from "../views/ConfirmationSentView/ConfirmationSentView";
import OneTimePasswordRegistrationView from "../views/OneTimePasswordRegistrationView/OneTimePasswordRegistrationView";
import SecurityKeyRegistrationView from "../views/SecurityKeyRegistrationView/SecurityKeyRegistrationView";
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
  path: '/confirm',
  title: 'e-mail sent',
  component: ConfirmationSent
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
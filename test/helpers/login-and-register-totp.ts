import VisitPage from "./visit-page";
import FillLoginPageAndClick from './fill-login-page-and-click';
import RegisterTotp from './register-totp';
import WaitRedirected from './wait-redirected';
import LoginAs from './login-as';

export default function(driver: any, user: string, email?: boolean) {
  return LoginAs(driver, user)
    .then(() => WaitRedirected(driver, "https://login.example.com:8080/secondfactor"))
    .then(() => RegisterTotp(driver, email));
}
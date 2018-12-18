import RegisterTotp from './register-totp';
import WaitRedirected from './wait-redirected';
import LoginAs from './login-as';
import Bluebird = require("bluebird");

export default function(driver: any, user: string, email?: boolean): Bluebird<string> {
  return LoginAs(driver, user)
    .then(() => WaitRedirected(driver, "https://login.example.com:8080/secondfactor"))
    .then(() => RegisterTotp(driver, email));
}
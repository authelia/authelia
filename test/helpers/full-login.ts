import VisitPage from "./visit-page";
import FillLoginPageWithUserAndPasswordAndClick from "./fill-login-page-and-click";
import ValidateTotp from "./validate-totp";
import WaitRedirected from "./wait-redirected";

// Validate the two factors!
export default function(driver: any, url: string, user: string, secret: string) {
  return VisitPage(driver, `https://login.example.com:8080/?rd=${url}`)
    .then(() => FillLoginPageWithUserAndPasswordAndClick(driver, user, 'password'))
    .then(() => ValidateTotp(driver, secret))
    .then(() => WaitRedirected(driver, "https://admin.example.com:8080/secret.html"));
}
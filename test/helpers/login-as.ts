import VisitPage from "./visit-page";
import FillLoginPageAndClick from './fill-login-page-and-click';

export default function(driver: any, user: string) {
  return VisitPage(driver, "https://login.example.com:8080/")
    .then(() => FillLoginPageAndClick(driver, user, "password"));
}
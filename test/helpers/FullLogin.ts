import VisitPage from "./VisitPage";
import FillLoginPageWithUserAndPasswordAndClick from "./FillLoginPageAndClick";
import ValidateTotp from "./ValidateTotp";
import WaitRedirected from "./WaitRedirected";
import { WebDriver } from "selenium-webdriver";

// Validate the two factors!
export default async function(driver: WebDriver, url: string, user: string, secret: string) {
  await VisitPage(driver, `https://login.example.com:8080/?rd=${url}`);
  await FillLoginPageWithUserAndPasswordAndClick(driver, user, 'password');
  await ValidateTotp(driver, secret);
  await WaitRedirected(driver, "https://admin.example.com:8080/secret.html");
}
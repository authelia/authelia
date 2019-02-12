import VisitPage from "./VisitPage";
import FillLoginPageAndClick from "./FillLoginPageAndClick";
import ValidateTotp from "./ValidateTotp";
import VerifyUrlIs from "./assertions/VerifyUrlIs";
import { WebDriver } from "selenium-webdriver";

// Validate the two factors!
export default async function(driver: WebDriver, user: string, secret: string, url: string) {
  await VisitPage(driver, `https://login.example.com:8080/?rd=${url}`);
  await FillLoginPageAndClick(driver, user, 'password');
  await ValidateTotp(driver, secret);
  await VerifyUrlIs(driver, url);
}
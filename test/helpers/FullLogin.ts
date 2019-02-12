import FillLoginPageAndClick from "./FillLoginPageAndClick";
import ValidateTotp from "./ValidateTotp";
import VerifyUrlIs from "./assertions/VerifyUrlIs";
import { WebDriver } from "selenium-webdriver";
import VisitPageAndWaitUrlIs from "./behaviors/VisitPageAndWaitUrlIs";

// Validate the two factors!
export default async function(driver: WebDriver, user: string, secret: string, url: string) {
  await VisitPageAndWaitUrlIs(driver, `https://login.example.com:8080/?rd=${url}`);
  await FillLoginPageAndClick(driver, user, 'password');
  await ValidateTotp(driver, secret);
  await VerifyUrlIs(driver, url);
}
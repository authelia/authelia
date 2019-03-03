import FillLoginPageAndClick from "./FillLoginPageAndClick";
import ValidateTotp from "./ValidateTotp";
import { WebDriver } from "selenium-webdriver";
import VisitPageAndWaitUrlIs from "./behaviors/VisitPageAndWaitUrlIs";

// Validate the two factors!
export default async function(driver: WebDriver, user: string, secret: string, url: string, timeout: number = 5000) {
  await VisitPageAndWaitUrlIs(driver, `https://login.example.com:8080/#/?rd=${url}`, timeout);
  await FillLoginPageAndClick(driver, user, 'password', false, timeout);
  await ValidateTotp(driver, secret, timeout);
}
import { WebDriver } from "selenium-webdriver";
import LoginAndRegisterTotp from "../LoginAndRegisterTotp";
import FullLogin from "../FullLogin";
import VisitPage from "../VisitPage";
import FillLoginPageAndClick from "../FillLoginPageAndClick";
import VerifyUrlIs from "../assertions/VerifyUrlIs";

export default async function(
  driver: WebDriver,
  username: string,
  password: string,
  targetUrl: string) {

    await VisitPage(driver, `https://login.example.com:8080/?rd=${targetUrl}`);
    await FillLoginPageAndClick(driver, username, password);
    await VerifyUrlIs(driver, targetUrl);
};
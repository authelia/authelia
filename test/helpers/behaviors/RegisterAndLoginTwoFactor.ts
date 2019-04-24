import { WebDriver } from "selenium-webdriver";
import LoginAndRegisterTotp from "../LoginAndRegisterTotp";
import FullLogin from "../FullLogin";
import VerifyUrlIs from "../assertions/VerifyUrlIs";
import VisitPageAndWaitUrlIs from "./VisitPageAndWaitUrlIs";
import ValidateTotp from "../ValidateTotp";
import FillLoginPageAndClick from "../FillLoginPageAndClick";

export default async function(
  driver: WebDriver,
  username: string,
  password: string,
  email: boolean = false,
  targetUrl: string = "https://login.example.com:8080/#/",
  timeout: number = 5000) {

  const secret = await LoginAndRegisterTotp(driver, username, password, email, timeout);
  await VisitPageAndWaitUrlIs(driver, `https://login.example.com:8080/#/?rd=${targetUrl}`, timeout);
  await ValidateTotp(driver, secret, timeout);
  await VerifyUrlIs(driver, targetUrl, timeout);
  return secret;
};
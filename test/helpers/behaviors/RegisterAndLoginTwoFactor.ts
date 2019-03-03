import { WebDriver } from "selenium-webdriver";
import LoginAndRegisterTotp from "../LoginAndRegisterTotp";
import FullLogin from "../FullLogin";
import VerifyUrlIs from "../assertions/VerifyUrlIs";

export default async function(
  driver: WebDriver,
  username: string,
  password: string,
  email: boolean = false,
  targetUrl: string = "https://login.example.com:8080/#/",
  timeout: number = 5000) {

  const secret = await LoginAndRegisterTotp(driver, username, password, email, timeout);
  await FullLogin(driver, username, secret, targetUrl, timeout);
  await VerifyUrlIs(driver, targetUrl, timeout);
  return secret;
};
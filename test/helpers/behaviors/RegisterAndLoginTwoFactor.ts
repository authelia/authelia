import { WebDriver } from "selenium-webdriver";
import LoginAndRegisterTotp from "../LoginAndRegisterTotp";
import FullLogin from "../FullLogin";

export default async function(
  driver: WebDriver,
  username: string,
  email: boolean = false,
  targetUrl: string = "https://login.example.com:8080/") {

  const secret = await LoginAndRegisterTotp(driver, username, email);
  await FullLogin(driver, username, secret, targetUrl);
};
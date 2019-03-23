import { WebDriver } from "selenium-webdriver";
import FillLoginPageAndClick from "../FillLoginPageAndClick";
import VerifyUrlIs from "../assertions/VerifyUrlIs";
import VisitPageAndWaitUrlIs from "./VisitPageAndWaitUrlIs";

export default async function(
  driver: WebDriver,
  username: string,
  password: string,
  targetUrl: string,
  timeout: number = 5000) {

  await VisitPageAndWaitUrlIs(driver, `https://login.example.com:8080/#/?rd=${targetUrl}`, timeout);
  await FillLoginPageAndClick(driver, username, password, false, timeout);
  await VerifyUrlIs(driver, targetUrl, timeout);
};
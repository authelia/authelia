import { WebDriver } from "selenium-webdriver";
import FillLoginPageAndClick from "../FillLoginPageAndClick";
import VerifyUrlIs from "../assertions/VerifyUrlIs";
import VisitPageAndWaitUrlIs from "./VisitPageAndWaitUrlIs";

export default async function(
  driver: WebDriver,
  username: string,
  password: string,
  targetUrl: string) {

    await VisitPageAndWaitUrlIs(driver, `https://login.example.com:8080/?rd=${targetUrl}`);
    await FillLoginPageAndClick(driver, username, password);
    await VerifyUrlIs(driver, targetUrl);
};
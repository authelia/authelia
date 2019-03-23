import FillLoginPageAndClick from './FillLoginPageAndClick';
import { WebDriver } from "selenium-webdriver";
import VisitPageAndWaitUrlIs from "./behaviors/VisitPageAndWaitUrlIs";

export default async function(driver: WebDriver, user: string, password: string, targetUrl?: string, timeout: number = 5000) {
  const urlExt = (targetUrl) ? ('rd=' + targetUrl) : '';
  await VisitPageAndWaitUrlIs(driver, "https://login.example.com:8080/#/?" + urlExt, timeout);
  await FillLoginPageAndClick(driver, user, password, false, timeout);
}
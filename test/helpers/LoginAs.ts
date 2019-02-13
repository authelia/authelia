import FillLoginPageAndClick from './FillLoginPageAndClick';
import { WebDriver } from "selenium-webdriver";
import VisitPageAndWaitUrlIs from "./behaviors/VisitPageAndWaitUrlIs";

export default async function(driver: WebDriver, user: string, password: string, targetUrl?: string) {
  const urlExt = (targetUrl) ? ('rd=' + targetUrl) : '';
  await VisitPageAndWaitUrlIs(driver, "https://login.example.com:8080/" + urlExt);
  await FillLoginPageAndClick(driver, user, password);
}
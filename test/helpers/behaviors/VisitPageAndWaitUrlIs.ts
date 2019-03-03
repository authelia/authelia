import { WebDriver } from "selenium-webdriver";
import VerifyUrlIs from "../assertions/VerifyUrlIs";
import VisitPage from "../VisitPage";

export default async function(driver: WebDriver, url: string, timeout: number = 5000) {
  await VisitPage(driver, url);
  await VerifyUrlIs(driver, url, timeout);
}
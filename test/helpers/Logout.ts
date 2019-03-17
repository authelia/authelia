import { WebDriver } from "selenium-webdriver";
import VisitPage from "./VisitPage";

export default async function(driver: WebDriver) {
  await VisitPage(driver, `https://login.example.com:8080/#/logout`);
}
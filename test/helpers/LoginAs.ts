import VisitPage from "./VisitPage";
import FillLoginPageAndClick from './FillLoginPageAndClick';
import { WebDriver } from "selenium-webdriver";

export default async function(driver: WebDriver, user: string, password: string = "password") {
  await VisitPage(driver, "https://login.example.com:8080/");
  await FillLoginPageAndClick(driver, user, password);
}
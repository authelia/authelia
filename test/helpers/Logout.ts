import { WebDriver } from "selenium-webdriver";
import VerifyIsFirstFactorStage from "./assertions/VerifyIsFirstFactorStage";
import VisitPage from "./VisitPage";

export default async function(driver: WebDriver) {
  await VisitPage(driver, "https://login.example.com:8080/#/logout");
  await VerifyIsFirstFactorStage(driver);
}
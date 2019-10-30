import { WebDriver } from "selenium-webdriver";
import VerifyIsFirstFactorStage from "./assertions/VerifyIsFirstFactorStage";
import VisitPage from "./VisitPage";
import VerifyUrlContains from "./assertions/VerifyUrlContains";

export default async function(driver: WebDriver) {
  await VisitPage(driver, "https://login.example.com:8080/#/logout");
  await VerifyUrlContains(driver, "https://login.example.com:8080/#/");
  await VerifyIsFirstFactorStage(driver);
}
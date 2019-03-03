import { WebDriver } from "selenium-webdriver";
import VerifyBodyContains from "./VerifyBodyContains";

// Verify if the current page contains "This is a very important secret!".
export default async function(driver: WebDriver, timeout: number = 5000) {
  await VerifyBodyContains(driver, "This is a very important secret!", timeout);
}
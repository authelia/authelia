import SeleniumWebDriver, { WebDriver } from "selenium-webdriver";
import VerifyElementDoesNotExist from "./VerifyElementDoesNotExist";

/**
 * Verify that an element does not exist.
 *
 * @param driver The selenium driver
 * @param content The content of the button to select.
 */
export default async function(driver: WebDriver, content: string) {
  await VerifyElementDoesNotExist(driver, SeleniumWebDriver.By.xpath("//button[text()='" + content + "']"))
}
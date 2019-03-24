import SeleniumWebDriver, { WebDriver } from "selenium-webdriver";
import VerifyElementExists from "./VerifyElementExists";

/**
 * Verify if a button with given content exists in the DOM.
 * @param driver The selenium web driver.
 * @param content The content of the button to find in the DOM.
 */
export default async function(driver: WebDriver, content: string) {
  await VerifyElementExists(driver, SeleniumWebDriver.By.xpath("//button[text()='" + content + "']"));
}
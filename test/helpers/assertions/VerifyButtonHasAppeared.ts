import SeleniumWebDriver, { WebDriver } from "selenium-webdriver";
import VerifyHasAppeared from "./VerifyHasAppeared";

/**
 * Verify if a button with given content exists in the DOM.
 * @param driver The selenium web driver.
 * @param content The content of the button to find in the DOM.
 */
export default async function(driver: WebDriver, content: string) {
  try {
    await VerifyHasAppeared(driver, SeleniumWebDriver.By.xpath("//button[text()='" + content + "']"));
  } catch (err) {
    throw new Error(`Button with content "${content}" should have appeared.`);
  }
}
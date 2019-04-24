import SeleniumWebdriver, { WebDriver, Key, ActionSequence } from "selenium-webdriver";

export default async function(driver: WebDriver, fieldId: string, timeout: number = 5000) {
    const element = await driver.wait(
        SeleniumWebdriver.until.elementLocated(
            SeleniumWebdriver.By.id(fieldId)), timeout)

    await element.sendKeys(Key.chord(Key.CONTROL, "a", Key.BACK_SPACE));
};
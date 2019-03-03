import { WebDriver } from "selenium-webdriver";

export default async function(driver: WebDriver, url: string) {
  await driver.get(url);
}
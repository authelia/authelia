import { WebDriver } from "selenium-webdriver";

export default async function(driver: WebDriver) {
  await driver.get(`https://login.example.com:8080/logout`);
}
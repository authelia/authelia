import SeleniumWebDriver from "selenium-webdriver"
import VisitPageAndWaitUrlIs from "./VisitPageAndWaitUrlIs";
import ClickOnLink from "../ClickOnLink";
import VerifyUrlIs from "../assertions/VerifyUrlIs";
import FillField from "../FillField";
import ClickOn from "../ClickOn";
import { GetLinkFromEmail } from "../GetIdentityLink";

export default async function(driver: SeleniumWebDriver.WebDriver, username: string, password: string, timeout: number = 5000) {
    await VisitPageAndWaitUrlIs(driver, "https://login.example.com:8080/#/");
    await ClickOnLink(driver, "Forgot password\?");
    await VerifyUrlIs(driver, "https://login.example.com:8080/#/forgot-password");
    await FillField(driver, "username", username);
    await ClickOn(driver, SeleniumWebDriver.By.id('next-button'));
    await VerifyUrlIs(driver, 'https://login.example.com:8080/#/confirmation-sent');

    await driver.sleep(500); // Simulate the time it takes to receive the e-mail.
    const link = await GetLinkFromEmail();
    await VisitPageAndWaitUrlIs(driver, link);
    await FillField(driver, "password1", password);
    await FillField(driver, "password2", password);
    await ClickOn(driver, SeleniumWebDriver.By.id('reset-button'));
}
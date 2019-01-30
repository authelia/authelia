import VisitPage from '../../helpers/VisitPage';
import ClickOnLink from '../../helpers/ClickOnLink';
import ClickOn from '../../helpers/ClickOn';
import WaitRedirected from '../../helpers/WaitRedirected';
import FillField from "../../helpers/FillField";
import {GetLinkFromEmail} from "../../helpers/GetIdentityLink";
import FillLoginPageAndClick from "../../helpers/FillLoginPageAndClick";
import SeleniumWebDriver from 'selenium-webdriver';
import IsSecondFactorStage from "../../helpers/IsSecondFactorStage";

export default function() {
  it("should reset password for john", async function() {
    await VisitPage(this.driver, "https://login.example.com:8080/");
    await ClickOnLink(this.driver, "Forgot password\?");
    await WaitRedirected(this.driver, "https://login.example.com:8080/forgot-password");
    await FillField(this.driver, "username", "john");
    await ClickOn(this.driver, SeleniumWebDriver.By.id('next-button'));

    await this.driver.sleep(500); // Simulate the time it takes to receive the e-mail.
    const link = await GetLinkFromEmail();
    await VisitPage(this.driver, link);
    await FillField(this.driver, "password1", "newpass");
    await FillField(this.driver, "password2", "newpass");
    await ClickOn(this.driver, SeleniumWebDriver.By.id('reset-button'));
    await WaitRedirected(this.driver, "https://login.example.com:8080/");
    await FillLoginPageAndClick(this.driver, "john", "newpass");
    await IsSecondFactorStage(this.driver);
  });
}

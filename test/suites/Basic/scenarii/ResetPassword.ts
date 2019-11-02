import SeleniumWebDriver from 'selenium-webdriver';

import ClickOnLink from '../../../helpers/ClickOnLink';
import ClickOn from '../../../helpers/ClickOn';
import FillField from "../../../helpers/FillField";
import {GetLinkFromEmail} from "../../../helpers/GetIdentityLink";
import FillLoginPageAndClick from "../../../helpers/FillLoginPageAndClick";
import IsSecondFactorStage from "../../../helpers/assertions/VerifyIsSecondFactorStage";
import VisitPageAndWaitUrlIs from '../../../helpers/behaviors/VisitPageAndWaitUrlIs';
import VerifyNotificationDisplayed from '../../../helpers/assertions/VerifyNotificationDisplayed';
import VerifyUrlIs from '../../../helpers/assertions/VerifyUrlIs';
import { StartDriver, StopDriver } from '../../../helpers/context/WithDriver';
import ResetPassword from '../../../helpers/behaviors/ResetPassword';

export default function() {
  beforeEach(async function() {
    this.driver = await StartDriver();
  });

  afterEach(async function() {
    await StopDriver(this.driver);
  })

  it("should reset password for john", async function() {
    await ResetPassword(this.driver, "john", "newpass");
    await VerifyUrlIs(this.driver, "https://login.example.com:8080/#/");
    await FillLoginPageAndClick(this.driver, "john", "newpass");

    // The user reaches the second factor page using the new password.
    await IsSecondFactorStage(this.driver);

    // restore password
    await ClickOnLink(this.driver, "Logout");
    await ResetPassword(this.driver, "john", "password");
  });

  it("should make attacker think reset password is initiated", async function() {
    await VisitPageAndWaitUrlIs(this.driver, "https://login.example.com:8080/#/");
    await ClickOnLink(this.driver, "Forgot password\?");
    await VerifyUrlIs(this.driver, "https://login.example.com:8080/#/forgot-password");
    await FillField(this.driver, "username", "unknown");
    await ClickOn(this.driver, SeleniumWebDriver.By.id('next-button'));

    // The malicious user thinks the confirmation has been sent.
    await VerifyUrlIs(this.driver, 'https://login.example.com:8080/#/confirmation-sent');
  });

  it("should notify passwords are different in reset form", async function() {
    await VisitPageAndWaitUrlIs(this.driver, "https://login.example.com:8080/#/");
    await ClickOnLink(this.driver, "Forgot password\?");
    await VerifyUrlIs(this.driver, "https://login.example.com:8080/#/forgot-password");
    await FillField(this.driver, "username", "john");
    await ClickOn(this.driver, SeleniumWebDriver.By.id('next-button'));
    await VerifyUrlIs(this.driver, 'https://login.example.com:8080/#/confirmation-sent');

    await this.driver.sleep(500); // Simulate the time it takes to receive the e-mail.
    const link = await GetLinkFromEmail();
    await VisitPageAndWaitUrlIs(this.driver, link);
    await FillField(this.driver, "password1", "newpass");
    await FillField(this.driver, "password2", "badpass");
    await ClickOn(this.driver, SeleniumWebDriver.By.id('reset-button'));
    await VerifyNotificationDisplayed(this.driver, "The passwords are different.");
  });
}

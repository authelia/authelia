import LoginAndRegisterTotp from "../../../helpers/LoginAndRegisterTotp";
import VisitPage from "../../../helpers/VisitPage";
import FillLoginPageWithUserAndPasswordAndClick from "../../../helpers/FillLoginPageAndClick";
import ValidateTotp from "../../../helpers/ValidateTotp";
import WaitRedirected from "../../../helpers/WaitRedirected";
import { WebDriver } from "selenium-webdriver";

export default function(this: Mocha.ISuiteCallbackContext) {
  this.timeout(20000);

  beforeEach(async function() {
    this.secret = await LoginAndRegisterTotp(this.driver, "john", true);
  });

  it("should disconnect user after inactivity period", async function() {
    const driver = this.driver as WebDriver;
    await VisitPage(driver, "https://login.example.com:8080/?rd=https://admin.example.com:8080/secret.html");
    await FillLoginPageWithUserAndPasswordAndClick(driver, 'john', 'password', false);
    await ValidateTotp(driver, this.secret);
    await WaitRedirected(driver, "https://admin.example.com:8080/secret.html");
    await VisitPage(driver, "https://home.example.com:8080/");
    await driver.sleep(6000);
    await driver.get("https://admin.example.com:8080/secret.html");
    await WaitRedirected(driver, "https://login.example.com:8080/?rd=https://admin.example.com:8080/secret.html");
  });

  it('should disconnect user after cookie expiration', async function() {
    const driver = this.driver as WebDriver;
    await VisitPage(driver, "https://login.example.com:8080/?rd=https://admin.example.com:8080/secret.html");
    await FillLoginPageWithUserAndPasswordAndClick(driver, 'john', 'password', false);
    await ValidateTotp(driver, this.secret);
    await WaitRedirected(driver, "https://admin.example.com:8080/secret.html");
    await VisitPage(driver, "https://home.example.com:8080/");

    await driver.sleep(4000);
    await driver.get("https://admin.example.com:8080/secret.html");
    await driver.sleep(2000);
    await driver.get("https://admin.example.com:8080/secret.html");

    await driver.sleep(2000);
    await driver.get("https://admin.example.com:8080/secret.html");
    await WaitRedirected(driver, "https://login.example.com:8080/?rd=https://admin.example.com:8080/secret.html");

  });

  describe('With remember me checkbox checked', function() {
    it("should keep user logged in after inactivity period", async function() {
      const driver = this.driver as WebDriver;
      await VisitPage(driver, "https://login.example.com:8080/?rd=https://admin.example.com:8080/secret.html");
      await FillLoginPageWithUserAndPasswordAndClick(driver, 'john', 'password', true);
      await ValidateTotp(driver, this.secret);
      await WaitRedirected(driver, "https://admin.example.com:8080/secret.html");
      await VisitPage(driver, "https://home.example.com:8080/");
      await driver.sleep(6000);
      await driver.get("https://admin.example.com:8080/secret.html");
      await WaitRedirected(driver, "https://admin.example.com:8080/secret.html");
    });
  });
}
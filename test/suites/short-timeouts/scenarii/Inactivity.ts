import LoginAndRegisterTotp from "../../../helpers/LoginAndRegisterTotp";
import FillLoginPageWithUserAndPasswordAndClick from "../../../helpers/FillLoginPageAndClick";
import ValidateTotp from "../../../helpers/ValidateTotp";
import { WebDriver } from "selenium-webdriver";
import VisitPageAndWaitUrlIs from "../../../helpers/behaviors/VisitPageAndWaitUrlIs";
import VisitPage from "../../../helpers/VisitPage";
import VerifyUrlIs from "../../../helpers/assertions/VerifyUrlIs";
import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";

export default function(this: Mocha.ISuiteCallbackContext) {
  this.timeout(20000);

  beforeEach(async function() {
    this.driver = await StartDriver();
    this.secret = await LoginAndRegisterTotp(this.driver, "john", "password", true);
  });

  afterEach(async function() {
    await StopDriver(this.driver);
  })

  it("should disconnect user after inactivity period", async function() {
    const driver = this.driver as WebDriver;
    await VisitPageAndWaitUrlIs(driver, "https://login.example.com:8080/?rd=https://admin.example.com:8080/secret.html");
    await FillLoginPageWithUserAndPasswordAndClick(driver, 'john', 'password', false);
    await ValidateTotp(driver, this.secret);
    await VerifyUrlIs(driver, "https://admin.example.com:8080/secret.html");
    await VisitPageAndWaitUrlIs(driver, "https://home.example.com:8080/");
    await driver.sleep(6000);
    await driver.get("https://admin.example.com:8080/secret.html");
    await VerifyUrlIs(driver, "https://login.example.com:8080/?rd=https://admin.example.com:8080/secret.html");
  });

  it('should disconnect user after cookie expiration', async function() {
    const driver = this.driver as WebDriver;
    await VisitPageAndWaitUrlIs(driver, "https://login.example.com:8080/?rd=https://admin.example.com:8080/secret.html");
    await FillLoginPageWithUserAndPasswordAndClick(driver, 'john', 'password', false);
    await ValidateTotp(driver, this.secret);
    await VerifyUrlIs(driver, "https://admin.example.com:8080/secret.html");
    await VisitPageAndWaitUrlIs(driver, "https://home.example.com:8080/");

    await driver.sleep(4000);
    await driver.get("https://admin.example.com:8080/secret.html");
    await driver.sleep(2000);
    await driver.get("https://admin.example.com:8080/secret.html");

    await driver.sleep(2000);
    await driver.get("https://admin.example.com:8080/secret.html");
    await VerifyUrlIs(driver, "https://login.example.com:8080/?rd=https://admin.example.com:8080/secret.html");

  });

  describe('With remember me checkbox checked', function() {
    it("should keep user logged in after inactivity period", async function() {
      const driver = this.driver as WebDriver;
      await VisitPageAndWaitUrlIs(driver, "https://login.example.com:8080/?rd=https://admin.example.com:8080/secret.html");
      await FillLoginPageWithUserAndPasswordAndClick(driver, 'john', 'password', true);
      await ValidateTotp(driver, this.secret);
      await VerifyUrlIs(driver, "https://admin.example.com:8080/secret.html");
      await VisitPageAndWaitUrlIs(driver, "https://home.example.com:8080/");
      await driver.sleep(6000);
      await VisitPage(driver, "https://admin.example.com:8080/secret.html");
      await VerifyUrlIs(driver, "https://admin.example.com:8080/secret.html");
    });
  });
}
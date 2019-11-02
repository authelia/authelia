import LoginAndRegisterTotp from "../../../helpers/LoginAndRegisterTotp";
import ValidateTotp from "../../../helpers/ValidateTotp";
import VisitPageAndWaitUrlIs from "../../../helpers/behaviors/VisitPageAndWaitUrlIs";
import VisitPage from "../../../helpers/VisitPage";
import VerifyUrlIs from "../../../helpers/assertions/VerifyUrlIs";
import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import ClickOnLink from "../../../helpers/ClickOnLink";
import FillLoginPageAndClick from "../../../helpers/FillLoginPageAndClick";

export default function(this: Mocha.ISuiteCallbackContext) {
  this.timeout(20000);

  beforeEach(async function() {
    this.driver = await StartDriver();
  });

  afterEach(async function() {
    await StopDriver(this.driver);
  })

  describe('Remember me not checked', function() {
    beforeEach(async function() {
      this.secret = await LoginAndRegisterTotp(this.driver, "john", "password", true);
    });

    it("should disconnect user after inactivity period", async function() {
      await VisitPageAndWaitUrlIs(this.driver, "https://login.example.com:8080/#/?rd=https://admin.example.com:8080/secret.html");
      await ValidateTotp(this.driver, this.secret);
      await VerifyUrlIs(this.driver, "https://admin.example.com:8080/secret.html");
      await VisitPageAndWaitUrlIs(this.driver, "https://home.example.com:8080/");
      await this.driver.sleep(6000);
      await VisitPage(this.driver, "https://admin.example.com:8080/secret.html");
      await VerifyUrlIs(this.driver, "https://login.example.com:8080/#/?rd=https://admin.example.com:8080/secret.html");
    });
  
    it('should disconnect user after cookie expiration', async function() {
      await VisitPageAndWaitUrlIs(this.driver, "https://login.example.com:8080/#/?rd=https://admin.example.com:8080/secret.html");
      await ValidateTotp(this.driver, this.secret);
      await VerifyUrlIs(this.driver, "https://admin.example.com:8080/secret.html");
  
      await this.driver.sleep(2000);
      await VisitPageAndWaitUrlIs(this.driver, "https://admin.example.com:8080/secret.html");
      await this.driver.sleep(2000);
      await VisitPageAndWaitUrlIs(this.driver, "https://admin.example.com:8080/secret.html");
      await this.driver.sleep(2000);
      await VisitPageAndWaitUrlIs(this.driver, "https://admin.example.com:8080/secret.html");
      await this.driver.sleep(2000);
      await VisitPage(this.driver, "https://admin.example.com:8080/secret.html");
      await VerifyUrlIs(this.driver, "https://login.example.com:8080/#/?rd=https://admin.example.com:8080/secret.html");
    });
  });

  describe('With remember me checkbox checked', function() {
    beforeEach(async function() {
      this.secret = await LoginAndRegisterTotp(this.driver, "john", "password", true);
      await VisitPageAndWaitUrlIs(this.driver, "https://login.example.com:8080/#/");
      await ClickOnLink(this.driver, "Logout");
    });

    it("should keep user logged in after inactivity period", async function() {
      await VisitPageAndWaitUrlIs(this.driver, "https://login.example.com:8080/#/?rd=https://admin.example.com:8080/secret.html");
      await FillLoginPageAndClick(this.driver, "john", "password", true);
      await ValidateTotp(this.driver, this.secret);
      await VerifyUrlIs(this.driver, "https://admin.example.com:8080/secret.html");
      await VisitPageAndWaitUrlIs(this.driver, "https://home.example.com:8080/#/");
      await this.driver.sleep(9000);
      await VisitPage(this.driver, "https://admin.example.com:8080/secret.html");
      await VerifyUrlIs(this.driver, "https://admin.example.com:8080/secret.html");
    });
  });
}
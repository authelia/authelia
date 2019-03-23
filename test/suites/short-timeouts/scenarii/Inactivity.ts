import LoginAndRegisterTotp from "../../../helpers/LoginAndRegisterTotp";
import FillLoginPageWithUserAndPasswordAndClick from "../../../helpers/FillLoginPageAndClick";
import ValidateTotp from "../../../helpers/ValidateTotp";
import VisitPageAndWaitUrlIs from "../../../helpers/behaviors/VisitPageAndWaitUrlIs";
import VisitPage from "../../../helpers/VisitPage";
import VerifyUrlIs from "../../../helpers/assertions/VerifyUrlIs";
import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import Logout from "../../../helpers/Logout";

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
    await VisitPageAndWaitUrlIs(this.driver, "https://login.example.com:8080/#/?rd=https://admin.example.com:8080/secret.html");
    await FillLoginPageWithUserAndPasswordAndClick(this.driver, 'john', 'password', false);
    await ValidateTotp(this.driver, this.secret);
    await VerifyUrlIs(this.driver, "https://admin.example.com:8080/secret.html");
    await VisitPageAndWaitUrlIs(this.driver, "https://home.example.com:8080/");
    await this.driver.sleep(6000);
    await VisitPage(this.driver, "https://admin.example.com:8080/secret.html");
    await VerifyUrlIs(this.driver, "https://login.example.com:8080/#/?rd=https://admin.example.com:8080/secret.html");
  });

  it('should disconnect user after cookie expiration', async function() {
    await VisitPageAndWaitUrlIs(this.driver, "https://login.example.com:8080/#/?rd=https://admin.example.com:8080/secret.html");
    await FillLoginPageWithUserAndPasswordAndClick(this.driver, 'john', 'password', false);
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

  describe('With remember me checkbox checked', function() {
    it("should keep user logged in after inactivity period", async function() {
      await VisitPageAndWaitUrlIs(this.driver, "https://login.example.com:8080/#/?rd=https://admin.example.com:8080/secret.html");
      await FillLoginPageWithUserAndPasswordAndClick(this.driver, 'john', 'password', true);
      await ValidateTotp(this.driver, this.secret);
      await VerifyUrlIs(this.driver, "https://admin.example.com:8080/secret.html");
      await VisitPageAndWaitUrlIs(this.driver, "https://home.example.com:8080/");
      await this.driver.sleep(9000);
      await VisitPage(this.driver, "https://admin.example.com:8080/secret.html");
      await VerifyUrlIs(this.driver, "https://admin.example.com:8080/secret.html");
    });
  });
}
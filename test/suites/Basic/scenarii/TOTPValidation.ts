import FillLoginPageWithUserAndPasswordAndClick from '../../../helpers/FillLoginPageAndClick';
import ValidateTotp from '../../../helpers/ValidateTotp';
import VerifySecretObserved from "../../../helpers/assertions/VerifySecretObserved";
import LoginAndRegisterTotp from '../../../helpers/LoginAndRegisterTotp';
import VisitPageAndWaitUrlIs from '../../../helpers/behaviors/VisitPageAndWaitUrlIs';
import VerifyNotificationDisplayed from '../../../helpers/assertions/VerifyNotificationDisplayed';
import VerifyUrlIs from '../../../helpers/assertions/VerifyUrlIs';
import { StartDriver, StopDriver } from '../../../helpers/context/WithDriver';

export default function() {
  /**
 * Given john has registered a TOTP secret,
 * When he validates the TOTP second factor,
 * Then he has access to secret page.
 */
  describe('Successfully pass second factor with TOTP', function() {
    before(async function() {
      this.driver = await StartDriver();
      const secret = await LoginAndRegisterTotp(this.driver, "john", "password", true);
      if (!secret) throw new Error('No secret!');
      
      await VisitPageAndWaitUrlIs(this.driver, "https://login.example.com:8080/#/?rd=https://admin.example.com:8080/secret.html");
      await ValidateTotp(this.driver, secret);
    });

    after(async function() {
      await StopDriver(this.driver);
    });

    it("should be automatically redirected to secret page", async function() {
      await VerifyUrlIs(this.driver, "https://admin.example.com:8080/secret.html");
    });

    it("should access the secret", async function() {
      await VerifySecretObserved(this.driver);
    });
  });

  /**
 * Given john has registered a TOTP secret,
 * When he fails the TOTP challenge,
 * Then he gets a notification message.
 */
  describe('Fail validation of second factor with TOTP', function() {
    before(async function() {
      this.driver = await StartDriver();
      await LoginAndRegisterTotp(this.driver, "john", "password", true);
      const BAD_TOKEN = "125478";
        
      await VisitPageAndWaitUrlIs(this.driver, "https://login.example.com:8080/#/?rd=https://admin.example.com:8080/secret.html");
      await ValidateTotp(this.driver, BAD_TOKEN);
    });

    after(async function() {
      await StopDriver(this.driver);
    });

    it("get a notification message", async function() {
      await VerifyNotificationDisplayed(this.driver, "Authentication failed, please retry later.");
    });
  });
}

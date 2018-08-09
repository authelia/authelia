import SeleniumWebdriver = require("selenium-webdriver");
import WithDriver from '../helpers/with-driver';
import LoginAndRegisterTotp from '../helpers/login-and-register-totp';

/** 
 * Given the user logs in as john,
 * When he register a TOTP token,
 * Then he reach a page containing the secret as string an qrcode
 */
describe('Registering TOTP', function() {
  this.timeout(10000);
  WithDriver();

  describe('successfully login as john', function() {
    before('register successfully', function() {
      this.timeout(10000);
      return LoginAndRegisterTotp(this.driver, "john");
    })

    it("should see generated qrcode", function() {
      this.driver.findElement(
          SeleniumWebdriver.By.id("qrcode"),
          5000);
    });

    it("should see generated secret", function() {
      this.driver.findElement(
          SeleniumWebdriver.By.id("secret"),
          5000);
    });
  });
});

import SeleniumWebdriver, { WebDriver } from "selenium-webdriver";
import Assert from 'assert';
import LoginAndRegisterTotp from '../../../helpers/LoginAndRegisterTotp';

/** 
 * Given the user logs in as john,
 * When he register a TOTP token,
 * Then he reach a page containing the secret as string an qrcode
 */
export default function() {
  describe('successfully login as john', function() {
    beforeEach('register successfully', async function() {
      this.timeout(10000);
      await LoginAndRegisterTotp(this.driver, "john", true);
    })

    it("should see generated qrcode", async function() {
      await this.driver.wait(
        SeleniumWebdriver.until.elementLocated(
        SeleniumWebdriver.By.className("qrcode")),
        5000);
    });

    it("should see generated secret", async function() {
      await this.driver.wait(
        SeleniumWebdriver.until.elementLocated(
        SeleniumWebdriver.By.className("base32-secret")),
        5000);
    });

    it("should have user and issuer in otp url", async function() {
      // this.timeout(100000);
      const el = await (this.driver as WebDriver).wait(
        SeleniumWebdriver.until.elementLocated(
          SeleniumWebdriver.By.className('otpauth-secret')), 5000);
          
      const otpauthUrl = await el.getAttribute('innerText');
      const label = 'john';
      const issuer = 'example.com';

      Assert(new RegExp(`^otpauth://totp/${label}\\?secret=[A-Z0-9]+&issuer=${issuer}$`).test(otpauthUrl));
    })
  });
};

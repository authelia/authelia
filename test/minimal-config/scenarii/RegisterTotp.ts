import SeleniumWebdriver from "selenium-webdriver";
import LoginAndRegisterTotp from '../../helpers/LoginAndRegisterTotp';

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
  });
};

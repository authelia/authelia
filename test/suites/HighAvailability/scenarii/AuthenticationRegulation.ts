import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import LoginAs from "../../../helpers/LoginAs";
import VerifyNotificationDisplayed from "../../../helpers/assertions/VerifyNotificationDisplayed";
import VerifyIsSecondFactorStage from "../../../helpers/assertions/VerifyIsSecondFactorStage";
import ClearFieldById from "../../../helpers/behaviors/ClearFieldById";
import {WebDriver} from "selenium-webdriver";

async function ClearAndLoginAgain(driver: WebDriver, username: string, password: string) {
  await ClearFieldById(driver, "username");
  await LoginAs(driver, username, password);
}

export default function() {
  describe('Authelia regulates authentications when a hacker is brute forcing', function() {
    this.timeout(30000);
    beforeEach(async function() {
      this.driver = await StartDriver();
    });

    afterEach(async function() {
      await StopDriver(this.driver);
    });

    it("should return an error message when providing correct credentials the 4th time.", async function() {
      await LoginAs(this.driver, "blackhat", "bad-password");
      await VerifyNotificationDisplayed(this.driver, "Authentication failed. Check your credentials.");
      await ClearAndLoginAgain(this.driver, "blackhat", "bad-password");
      await VerifyNotificationDisplayed(this.driver, "Authentication failed. Check your credentials.");
      await ClearAndLoginAgain(this.driver, "blackhat", "bad-password");
      await VerifyNotificationDisplayed(this.driver, "Authentication failed. Check your credentials.");

      // when providing good credentials, the hacker is regulated and see same message as previously.
      await ClearAndLoginAgain(this.driver, "blackhat", "bad-password");
      await VerifyNotificationDisplayed(this.driver, "Authentication failed. Check your credentials.");

      // Wait the regulation ban time before retrying with correct credentials.
      // It should authenticate normally.
      await this.driver.sleep(12000);
      await ClearAndLoginAgain(this.driver, "blackhat", "password");
      await VerifyIsSecondFactorStage(this.driver);
    });
  });
}
import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import LoginAs from "../../../helpers/LoginAs";
import VerifyNotificationDisplayed from "../../../helpers/assertions/VerifyNotificationDisplayed";
import VerifyIsSecondFactorStage from "../../../helpers/assertions/VerifyIsSecondFactorStage";

/*
Given I visit "https://login.example.com:8080/"
    And I set field "username" to "blackhat"
    And I set field "password" to "bad-password"
    And I click on "Sign in"
    And I get a notification of type "error" with message "Authentication failed. Please check your credentials."
    And I set field "password" to "bad-password"
    And I click on "Sign in"
    And I get a notification of type "error" with message "Authentication failed. Please check your credentials."
    And I set field "password" to "bad-password"
    And I click on "Sign in"
    And I get a notification of type "error" with message "Authentication failed. Please check your credentials."
    When I set field "password" to "password"
    And I click on "Sign in"
    Then I get a notification of type "error" with message "Authentication failed. Please check your credentials."
*/

export default function() {
  describe('Authelia regulates authentications when a hacker is brute forcing', function() {
    this.timeout(15000);
    before(async function() {
      this.driver = await StartDriver();
    });

    after(async function() {
      await StopDriver(this.driver);
    });

    it("should return an error message when providing correct credentials the 4th time.", async function() {
      await LoginAs(this.driver, "blackhat", "bad-password");
      await VerifyNotificationDisplayed(this.driver, "Authentication failed. Please check your credentials.");
      await LoginAs(this.driver, "blackhat", "bad-password");
      await VerifyNotificationDisplayed(this.driver, "Authentication failed. Please check your credentials.");
      await LoginAs(this.driver, "blackhat", "bad-password");
      await VerifyNotificationDisplayed(this.driver, "Authentication failed. Please check your credentials.");

      // when providing good credentials, the hacker is regulated and see same message as previously.
      await LoginAs(this.driver, "blackhat", "password");
      await VerifyNotificationDisplayed(this.driver, "Authentication failed. Please check your credentials.");

      // Wait the regulation ban time before retrying with correct credentials.
      // It should authenticate normally.
      await this.driver.sleep(6000);
      await LoginAs(this.driver, "blackhat", "password");
      await VerifyIsSecondFactorStage(this.driver);
    });
  });
}
import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import LoginAs from "../../../helpers/LoginAs";
import VerifyIsOneTimePasswordView from "../../../helpers/assertions/VerifyIsOneTimePasswordView";
import ClickOnLink from "../../../helpers/ClickOnLink";
import VerifyIsUseAnotherMethodView from "../../../helpers/assertions/VerifyIsUseAnotherMethodView";
import ClickOnButton from "../../../helpers/behaviors/ClickOnButton";
import VerifyIsSecondFactorStage from "../../../helpers/assertions/VerifyIsSecondFactorStage";
import VerifyIsDuoPushNotificationView from "../../../helpers/assertions/VerifyIsDuoPushNotificationView";


// This fixture tests that the latest used method is still used when the user gets back.
export default function() {
  before(async function() {
    this.driver = await StartDriver();
  });

  after(async function() {
    await StopDriver(this.driver);
  });

  // The default method is TOTP and then everytime the user switches method,
  // it get remembered and reloaded during next authentication.
  it('should serve the correct method', async function() {
    await LoginAs(this.driver, "john", "password", "https://secure.example.com:8080/");
    await VerifyIsSecondFactorStage(this.driver);

    await ClickOnLink(this.driver, 'Use another method');
    await VerifyIsUseAnotherMethodView(this.driver);
    await ClickOnButton(this.driver, 'Duo Push Notification');
    
    // Verify that the user is redirected to the new method
    await VerifyIsDuoPushNotificationView(this.driver);
    await ClickOnLink(this.driver, "Logout");

    // Login with another user to check that he gets TOTP view.
    await LoginAs(this.driver, "harry", "password", "https://secure.example.com:8080/");
    await VerifyIsOneTimePasswordView(this.driver);
    await ClickOnLink(this.driver, "Logout");

    // Log john again to check that the prefered method has been persisted
    await LoginAs(this.driver, "john", "password", "https://secure.example.com:8080/");
    await VerifyIsDuoPushNotificationView(this.driver);

    // Restore the prefered method to one-time password.
    await ClickOnLink(this.driver, 'Use another method');
    await VerifyIsUseAnotherMethodView(this.driver);
    await ClickOnButton(this.driver, 'One-Time Password');
    await VerifyIsOneTimePasswordView(this.driver);
    await ClickOnLink(this.driver, "Logout");
  });
}
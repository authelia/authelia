import { StartDriver, StopDriver } from "../../helpers/context/WithDriver";
import LoginAs from "../../helpers/LoginAs";
import VerifyNotificationDisplayed from "../../helpers/assertions/VerifyNotificationDisplayed";
import VerifyIsSecondFactorStage from "../../helpers/assertions/VerifyIsSecondFactorStage";
import ClearFieldById from "../behaviors/ClearFieldById";

export default function(regulationMilliseconds: number) {
    return function () {
        describe('Authelia regulates authentications when a hacker is brute forcing', function() {
            this.timeout(30000);
            beforeEach(async function() {
                this.driver = await StartDriver();
            });
        
            afterEach(async function() {
                await StopDriver(this.driver);
            });
        
            it("should return an error message when providing correct credentials the 4th time.", async function() {
                await LoginAs(this.driver, "james", "bad-password");
                await VerifyNotificationDisplayed(this.driver, "Authentication failed. Check your credentials.");
                await ClearFieldById(this.driver, "username");

                await LoginAs(this.driver, "james", "bad-password");
                await VerifyNotificationDisplayed(this.driver, "Authentication failed. Check your credentials.");
                await ClearFieldById(this.driver, "username");

                await LoginAs(this.driver, "james", "bad-password");
                await VerifyNotificationDisplayed(this.driver, "Authentication failed. Check your credentials.");
                await ClearFieldById(this.driver, "username");

                // when providing good credentials, the hacker is regulated and see same message as previously.
                await LoginAs(this.driver, "james", "bad-password");
                await VerifyNotificationDisplayed(this.driver, "Authentication failed. Check your credentials.");
                await ClearFieldById(this.driver, "username");

                // Wait the regulation ban time before retrying with correct credentials.
                // It should authenticate normally.
                await this.driver.sleep(regulationMilliseconds);
                await LoginAs(this.driver, "james", "password");
                await VerifyIsSecondFactorStage(this.driver);
            });
        });
    }
}
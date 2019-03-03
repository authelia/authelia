import Logout from "../Logout";
import { StartDriver, StopDriver } from "../context/WithDriver";
import LoginOneFactor from "../behaviors/LoginOneFactor";
import VerifySecretObserved from "../assertions/VerifySecretObserved";
import VisitPage from "../VisitPage";
import VerifyIsSecondFactorStage from "../assertions/VerifyIsSecondFactorStage";
import VerifyUrlContains from "../assertions/VerifyUrlContains";

/*
 * Those tests are related to single factor protected resources.
 */
export default function(timeout: number = 5000) {
  return function() {
    beforeEach(async function() {
      this.driver = await StartDriver();
    });
  
    afterEach(async function() {
      await Logout(this.driver);
      await StopDriver(this.driver);
    });
  
    it("should redirect user after first stage", async function() {
      await LoginOneFactor(this.driver, "john", "password", "https://singlefactor.example.com:8080/secret.html", timeout);
      await VerifySecretObserved(this.driver, timeout);
    });
  
    it("should redirect to portal if not enough authorized", async function() {
      await LoginOneFactor(this.driver, "john", "password", "https://singlefactor.example.com:8080/secret.html", timeout);
      await VisitPage(this.driver, "https://admin.example.com:8080/secret.html");
  
      // the url should be the one from the portal.
      await VerifyUrlContains(this.driver, "https://login.example.com:8080/#/?rd=https://admin.example.com:8080", timeout);
  
      // And the user should end up on the second factor page.
      await VerifyIsSecondFactorStage(this.driver, timeout);
    });
  }
}
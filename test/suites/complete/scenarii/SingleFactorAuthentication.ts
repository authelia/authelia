import Logout from "../../../helpers/Logout";
import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import LoginOneFactor from "../../../helpers/behaviors/LoginOneFactor";
import VerifySecretObserved from "../../../helpers/assertions/VerifySecretObserved";
import VisitPage from "../../../helpers/VisitPage";
import VerifyUrlIs from "../../../helpers/assertions/VerifyUrlIs";
import VerifyIsSecondFactorStage from "../../../helpers/assertions/VerifyIsSecondFactorStage";

export default function() {
  beforeEach(async function() {
    this.driver = await StartDriver();
  });

  afterEach(async function() {
    await Logout(this.driver);
    await StopDriver(this.driver);
  });

  it("should redirect user after first stage", async function() {
    await LoginOneFactor(this.driver, "john", "password", "https://single_factor.example.com:8080/secret.html");
    await VerifySecretObserved(this.driver);
  });

  it("should redirect to portal if not enough authorized", async function() {
    await LoginOneFactor(this.driver, "john", "password", "https://single_factor.example.com:8080/secret.html");
    await VisitPage(this.driver, "https://admin.example.com:8080/secret.html");

    // the url should be the one from the portal.
    await VerifyUrlIs(this.driver, "https://login.example.com:8080/?rd=https://admin.example.com:8080/secret.html");

    // And the user should end up on the second factor page.
    await VerifyIsSecondFactorStage(this.driver);
  })
}
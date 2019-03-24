import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import LoginAs from "../../../helpers/LoginAs";
import VerifyIsSecondFactorStage from "../../../helpers/assertions/VerifyIsSecondFactorStage";
import ClickOnLink from "../../../helpers/ClickOnLink";
import VerifyIsUseAnotherMethodView from "../../../helpers/assertions/VerifyIsUseAnotherMethodView";
import VerifyElementDoesNotExist from "../../../helpers/assertions/VerifyElementDoesNotExist";
import SeleniumWebDriver from "selenium-webdriver";
import VerifyButtonDoesNotExist from "../../../helpers/assertions/VerifyButtonDoesNotExist";
import VerifyButtonExists from "../../../helpers/assertions/VerifyButtonExists";



export default function() {
  before(async function() {
    this.driver = await StartDriver();
  });

  after(async function() {
    await StopDriver(this.driver);
  });

  // The Duo API is not configured so we should not see the method.
  it("should not display duo push notification method", async function() {
    await LoginAs(this.driver, "john", "password", "https://secure.example.com:8080/");
    await VerifyIsSecondFactorStage(this.driver);

    await ClickOnLink(this.driver, 'Use another method');
    await VerifyIsUseAnotherMethodView(this.driver);
    await VerifyButtonExists(this.driver, "Security Key (U2F)");
    await VerifyButtonExists(this.driver, "One-Time Password");
    await VerifyButtonDoesNotExist(this.driver, "Duo Push Notification");
  });
}
import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import LoginAs from "../../../helpers/LoginAs";
import VerifyIsSecondFactorStage from "../../../helpers/assertions/VerifyIsSecondFactorStage";
import ClickOnLink from "../../../helpers/ClickOnLink";
import VerifyIsUseAnotherMethodView from "../../../helpers/assertions/VerifyIsUseAnotherMethodView";
import ClickOnButton from "../../../helpers/behaviors/ClickOnButton";
import Request from 'request-promise';
import VerifyIsAlreadyAuthenticatedStage from "../../../helpers/assertions/VerifyIsAlreadyAuthenticatedStage";

export default function() {
  before(async function() {
    this.driver = await StartDriver();

    // Configure the fake API to return allowing response.
    await Request('https://duo.example.com/allow', {method: 'POST'});
  });

  after(async function () {
    await StopDriver(this.driver);
  });

  it('should send user to already authenticated page', async function() {
    await LoginAs(this.driver, "john", "password");
    await VerifyIsSecondFactorStage(this.driver);

    await ClickOnLink(this.driver, 'Use another method');
    await VerifyIsUseAnotherMethodView(this.driver);
    await ClickOnButton(this.driver, 'Duo Push Notification');
    await VerifyIsAlreadyAuthenticatedStage(this.driver, 10000);

    await ClickOnButton(this.driver, "Logout");
  });
}
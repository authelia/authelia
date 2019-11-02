import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import LoginAs from "../../../helpers/LoginAs";
import VerifyIsSecondFactorStage from "../../../helpers/assertions/VerifyIsSecondFactorStage";
import ClickOnLink from "../../../helpers/ClickOnLink";
import VerifyIsUseAnotherMethodView from "../../../helpers/assertions/VerifyIsUseAnotherMethodView";
import ClickOnButton from "../../../helpers/behaviors/ClickOnButton";
import VerifySecretObserved from "../../../helpers/assertions/VerifySecretObserved";
import Request from 'request-promise';
import VerifyUrlIs from "../../../helpers/assertions/VerifyUrlIs";
import VerifyHasAppeared from "../../../helpers/assertions/VerifyHasAppeared";
import SeleniumWebDriver from "selenium-webdriver";
import VisitPage from "../../../helpers/VisitPage";


export default function() {
  before(async function() {
    this.driver = await StartDriver();
  });

  after(async function () {
    await StopDriver(this.driver);
  });

  describe('Allow access', function() {
    before(async function() {
      // Configure the fake API to return allowing response.
      await Request('https://duo.example.com/allow', {method: 'POST'});
    });

    it('should grant access with Duo API', async function() {
      await LoginAs(this.driver, "john", "password", "https://secure.example.com:8080/secret.html");
      await VerifyIsSecondFactorStage(this.driver);
  
      await ClickOnLink(this.driver, 'Use another method');
      await VerifyIsUseAnotherMethodView(this.driver);
      await ClickOnButton(this.driver, 'Duo Push Notification');
  
      await VerifyUrlIs(this.driver, "https://secure.example.com:8080/secret.html");
      await VerifySecretObserved(this.driver);

      await VisitPage(this.driver, "https://login.example.com:8080/#/");
      await ClickOnButton(this.driver, "Logout");
    });
  });

  describe('Deny access', function() {
    before(async function() {
      // Configure the fake API to return denying response.
      await Request('https://duo.example.com/deny', {method: 'POST'});
    });

    it('should grant access with Duo API', async function() {
      await LoginAs(this.driver, "john", "password", "https://secure.example.com:8080/secret.html");
      await VerifyIsSecondFactorStage(this.driver);
  
      await ClickOnLink(this.driver, 'Use another method');
      await VerifyIsUseAnotherMethodView(this.driver);
      await ClickOnButton(this.driver, 'Duo Push Notification');

      // The retry button appeared.
      await VerifyHasAppeared(this.driver, SeleniumWebDriver.By.tagName("button"));
    });
  });
}
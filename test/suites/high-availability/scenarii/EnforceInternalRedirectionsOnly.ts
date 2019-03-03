import LoginAndRegisterTotp from "../../../helpers/LoginAndRegisterTotp";
import FillLoginPageWithUserAndPasswordAndClick from '../../../helpers/FillLoginPageAndClick';
import ValidateTotp from "../../../helpers/ValidateTotp";
import Logout from "../../../helpers/Logout";
import VerifyIsAlreadyAuthenticatedStage from "../../../helpers/assertions/VerifyIsAlreadyAuthenticatedStage";
import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import VisitPageAndWaitUrlIs from "../../../helpers/behaviors/VisitPageAndWaitUrlIs";
import VerifyUrlIs from "../../../helpers/assertions/VerifyUrlIs";
import VerifyNotificationDisplayed from "../../../helpers/assertions/VerifyNotificationDisplayed";

/*
 * Authelia should not be vulnerable to open redirection. Otherwise it would aid an
 * attacker in conducting a phishing attack.
 * 
 * To avoid the issue, Authelia's client scans the URL and prevent any redirection if
 * the URL is pointing to an external domain.
 */
export default function() {
  describe("Only redirection to a subdomain of the protected domain should be allowed", function() {
    this.timeout(10000);
    let secret: string;
  
    beforeEach(async function() {
      this.driver = await StartDriver();
      secret = await LoginAndRegisterTotp(this.driver, "john", "password", true)
    });
  
    afterEach(async function() {
      await Logout(this.driver);
      await StopDriver(this.driver);
    })
  
    function CannotRedirectTo(url: string, twoFactor: boolean = true) {
      it(`should redirect to already authenticated page when requesting ${url}`, async function() {
        await VisitPageAndWaitUrlIs(this.driver, `https://login.example.com:8080/#/?rd=${url}`);
        await FillLoginPageWithUserAndPasswordAndClick(this.driver, 'john', 'password');
        if (twoFactor) {
          await ValidateTotp(this.driver, secret);
          await VerifyIsAlreadyAuthenticatedStage(this.driver);
        } else {
          await VerifyNotificationDisplayed(this.driver,
            "You're authenticated but cannot be automatically redirected to an unsafe URL.");
        }
      });
    }

    function CanRedirectTo(url: string) {
      it(`should redirect to ${url}`, async function() {
        await VisitPageAndWaitUrlIs(this.driver, `https://login.example.com:8080/#/?rd=${url}`);
        await FillLoginPageWithUserAndPasswordAndClick(this.driver, 'john', 'password');
        await ValidateTotp(this.driver, secret);
        await VerifyUrlIs(this.driver, url);
      });
    }
    
    describe('Cannot redirect to https://www.google.fr', function() {
      // Do not redirect to another domain than example.com
      CannotRedirectTo("https://www.google.fr");
    });

    describe('Cannot redirect to https://public.example.com.a:8080', function() {
      // Do not redirect to another domain than example.com
      CannotRedirectTo("https://public.example.com.a:8080");
    });

    describe('Cannot redirect to http://secure.example.com:8080', function() {
      // Do not redirect to http website
      CannotRedirectTo("http://secure.example.com:8080");
    });

    describe('Cannot redirect to http://singlefactor.example.com:8080', function() {
      // Do not redirect to http website
      CannotRedirectTo("http://singlefactor.example.com:8080", false);
    });

    describe('Can redirect to https://secure.example.com:8080/', function() {
      // Can redirect to any subdomain of the domain protected by Authelia.
      CanRedirectTo("https://secure.example.com:8080/");
    });
  });
}
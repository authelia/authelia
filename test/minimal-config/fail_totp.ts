require("chromedriver");
import WithDriver from '../helpers/with-driver';
import FillLoginPageWithUserAndPasswordAndClick from '../helpers/fill-login-page-and-click';
import VisitPage from '../helpers/visit-page';
import ValidateTotp from '../helpers/validate-totp';
import LoginAndRegisterTotp from '../helpers/login-and-register-totp';
import seeNotification from "../helpers/see-notification";
import {AUTHENTICATION_TOTP_FAILED} from '../../shared/UserMessages';

/**
 * Given john has registered a TOTP secret,
 * When he fails the TOTP challenge,
 * Then he gets a notification message.
 */
describe('Fail TOTP challenge', function() {
  this.timeout(10000);
  WithDriver();

  describe('successfully login as john', function() {
    before(function() {
      return LoginAndRegisterTotp(this.driver, "john", true);
    });

    describe('fail second factor', function() {
      before(function() {
        const BAD_TOKEN = "125478";
        const driver = this.driver;
        
        return VisitPage(driver, "https://login.example.com:8080/?rd=https://admin.example.com:8080/secret.html")
          .then(() => FillLoginPageWithUserAndPasswordAndClick(driver, 'john', 'password'))
          .then(() => ValidateTotp(driver, BAD_TOKEN));
      });

      it("get a notification message", function() {
        return seeNotification(this.driver, "error", AUTHENTICATION_TOTP_FAILED);
      });
    });
  });
});

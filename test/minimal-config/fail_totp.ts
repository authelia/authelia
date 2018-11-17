require("chromedriver");
import Bluebird = require("bluebird");
import SeleniumWebdriver = require("selenium-webdriver");
import Fs = require("fs");
import Speakeasy = require("speakeasy");
import WithDriver from '../helpers/with-driver';
import FillLoginPageWithUserAndPasswordAndClick from '../helpers/fill-login-page-and-click';
import WaitRedirected from '../helpers/wait-redirected';
import VisitPage from '../helpers/visit-page';
import RegisterTotp from '../helpers/register-totp';
import ValidateTotp from '../helpers/validate-totp';
import AccessSecret from "../helpers/access-secret";
import LoginAndRegisterTotp from '../helpers/login-and-register-totp';
import seeNotification from "../helpers/see-notification";

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
      const that = this;
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
        return seeNotification(this.driver, "error", "Authentication failed. Have you already registered your secret?");
      });
    });
  });
});

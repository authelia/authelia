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

/**
 * Given john has registered a TOTP secret,
 * When he validates the TOTP second factor,
 * Then he has access to secret page.
 */
describe('Validate TOTP factor', function() {
  this.timeout(10000);
  WithDriver();

  describe('successfully login as john', function() {
    before(function() {
      const that = this;
      return LoginAndRegisterTotp(this.driver, "john", true)
        .then(function(secret: string) {
          that.secret = secret;
        })
    });

    describe('validate second factor', function() {
      before(function() {
        const secret = this.secret;
        if(!secret) return Bluebird.reject(new Error("No secret!"));
        const driver = this.driver;
        
        return VisitPage(driver, "https://login.example.com:8080/?rd=https://admin.example.com:8080/secret.html")
          .then(() => {
            return FillLoginPageWithUserAndPasswordAndClick(driver, 'john', 'password');
          })
          .then(() => {
            return ValidateTotp(driver, secret);
          })
          .then(() => {
            return WaitRedirected(driver, "https://admin.example.com:8080/secret.html")
          });
      });

      it("should access the secret", function() {
        return AccessSecret(this.driver);
      });
    });
  });
});

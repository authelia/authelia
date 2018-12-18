require("chromedriver");
import Bluebird = require("bluebird");
import ChildProcess = require("child_process");

import WithDriver from '../helpers/with-driver';
import VisitPage from '../helpers/visit-page';
import ClickOnLink from '../helpers/click-on-link';
import ClickOnButton from '../helpers/click-on-button';
import WaitRedirect from '../helpers/wait-redirected';
import FillField from "../helpers/fill-field";
import {GetLinkFromEmail} from "../helpers/get-identity-link";
import FillLoginPageAndClick from "../helpers/fill-login-page-and-click";

const execAsync = Bluebird.promisify(ChildProcess.exec);

describe('Reset password', function() {
  this.timeout(10000);
  WithDriver();

  after(() => {
    return execAsync("cp users_database.yml users_database.test.yml");
  })

  describe('click on reset password', function() {
    it("should reset password for john", function() {
      return VisitPage(this.driver, "https://login.example.com:8080/")
        .then(() => ClickOnLink(this.driver, "Forgot password\?"))
        .then(() => WaitRedirect(this.driver, "https://login.example.com:8080/password-reset/request"))
        .then(() => FillField(this.driver, "username", "john"))
        .then(() => ClickOnButton(this.driver, "Reset Password"))
        .then(() => this.driver.sleep(1000)) // Simulate the time to read it from mailbox.
        .then(() => GetLinkFromEmail())
        .then((link) => VisitPage(this.driver, link))
        .then(() => FillField(this.driver, "password1", "newpass"))
        .then(() => FillField(this.driver, "password2", "newpass"))
        .then(() => ClickOnButton(this.driver, "Reset Password"))
        .then(() => WaitRedirect(this.driver, "https://login.example.com:8080/"))
        .then(() => FillLoginPageAndClick(this.driver, "john", "newpass"))
        .then(() => WaitRedirect(this.driver, "https://login.example.com:8080/secondfactor"))
    });
  });
});

require("chromedriver");
import SeleniumWebdriver = require("selenium-webdriver");
import WithDriver from '../helpers/with-driver';
import LoginAndRegisterTotp from '../helpers/login-and-register-totp';
import LoginAs from '../helpers/login-as';
import VisitPage from '../helpers/visit-page';

describe('Connection retry when mongo fails or restarts', function() {
  this.timeout(20000);
  WithDriver();

  it('should be able to login after mongo restarts', function() {
    const that = this;
    return that.environment.stop_service("mongo")
      .then(() => that.environment.restart_service("authelia", 2000))
      .then(() => that.environment.restart_service("mongo"))
      .then(() => LoginAs(that.driver, "john"));
  })
});

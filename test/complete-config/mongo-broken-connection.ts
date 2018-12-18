import WithDriver from '../helpers/with-driver';
import fullLogin from '../helpers/full-login';
import loginAndRegisterTotp from '../helpers/login-and-register-totp';

describe("Connection retry when mongo fails or restarts", function() {
  this.timeout(30000);
  WithDriver();

  it("should be able to login after mongo restarts", function() {
    const that = this;
    let secret;
    return loginAndRegisterTotp(that.driver, "john", true)
      .then(_secret => secret = _secret)
      .then(() => that.environment.restart_service("mongo", 1000))
      .then(() => fullLogin(that.driver, "https://admin.example.com:8080/secret.html", "john", secret));
  })
});

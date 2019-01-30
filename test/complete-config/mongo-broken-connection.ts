import WithDriver from '../helpers/context/WithDriver';
import fullLogin from '../helpers/FullLogin';
import loginAndRegisterTotp from '../helpers/LoginAndRegisterTotp';

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

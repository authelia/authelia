import WithDriver from "../helpers/context/WithDriver";
import LoginAndRegisterTotp from "../helpers/LoginAndRegisterTotp";
import SeeNotification from "../helpers/SeeNotification";
import VisitPage from "../helpers/VisitPage";
import FillLoginPageWithUserAndPasswordAndClick from '../helpers/FillLoginPageAndClick';
import ValidateTotp from "../helpers/ValidateTotp";
import {CANNOT_REDIRECT_TO_EXTERNAL_DOMAIN} from '../../shared/UserMessages';

/*
 * Authelia should not be vulnerable to open redirection. Otherwise it would aid an
 * attacker in conducting a phishing attack.
 * 
 * To avoid the issue, Authelia's client scans the URL and prevent any redirection if
 * the URL is pointing to an external domain.
 */
describe("Redirection should be performed only if in domain", function() {
  this.timeout(10000);
  WithDriver();

  before(function() {
    const that = this;
    return LoginAndRegisterTotp(this.driver, "john", true)
      .then((secret: string) => that.secret = secret)
  });

  function DoNotRedirect(url: string) {
    it(`should see an error message instead of redirecting to ${url}`, function() {
      const driver = this.driver;
      const secret = this.secret;
      return VisitPage(driver, `https://login.example.com:8080/?rd=${url}`)
        .then(() => FillLoginPageWithUserAndPasswordAndClick(driver, 'john', 'password'))
        .then(() => ValidateTotp(driver, secret))
        .then(() => SeeNotification(driver, "error", CANNOT_REDIRECT_TO_EXTERNAL_DOMAIN))
        .then(() => driver.get(`https://login.example.com:8080/logout`));
    });
  }

  DoNotRedirect("www.google.fr");
  DoNotRedirect("http://www.google.fr");
  DoNotRedirect("https://www.google.fr");
})
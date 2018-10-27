import WithDriver from '../helpers/with-driver';
import FillLoginPageWithUserAndPasswordAndClick from '../helpers/fill-login-page-and-click';
import VisitPage from '../helpers/visit-page';
import SeeNotification from '../helpers/see-notification';
import {AUTHENTICATION_FAILED} from '../../shared/UserMessages';

/**
 * When user provides bad password,
 * Then he gets a notification message.
 */
describe("Provide bad password", function() {
  WithDriver();

  describe('failed login as john', function() {
    before(function() {
      this.timeout(10000);

      const driver = this.driver;
      return VisitPage(driver, "https://login.example.com:8080/")
        .then(function() {
          return FillLoginPageWithUserAndPasswordAndClick(driver, 'john', 'bad_password');
        });
    });

    it('should get a notification message', function() {
      this.timeout(10000);
      return SeeNotification(this.driver, "error", AUTHENTICATION_FAILED);
    });
  });
});

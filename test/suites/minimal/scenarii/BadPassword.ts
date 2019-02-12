import FillLoginPageWithUserAndPasswordAndClick from '../../../helpers/FillLoginPageAndClick';
import SeeNotification from '../../../helpers/SeeNotification';
import {AUTHENTICATION_FAILED} from '../../../../shared/UserMessages';
import VisitPageAndWaitUrlIs from '../../../helpers/behaviors/VisitPageAndWaitUrlIs';

export default function() {
/**
 * When user provides bad password,
 * Then he gets a notification message.
 */
  describe('failed login as john in first factor', function() {
    beforeEach(async function() {
      this.timeout(10000);
      await VisitPageAndWaitUrlIs(this.driver, "https://login.example.com:8080/")
      await FillLoginPageWithUserAndPasswordAndClick(this.driver, 'john', 'bad_password');
    });

    it('should get a notification message', async function () {
      this.timeout(10000);
      await SeeNotification(this.driver, "error", AUTHENTICATION_FAILED);
    });
  });
}

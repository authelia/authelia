import FillLoginPageWithUserAndPasswordAndClick from '../../../helpers/FillLoginPageAndClick';
import VisitPageAndWaitUrlIs from '../../../helpers/behaviors/VisitPageAndWaitUrlIs';
import VerifyNotificationDisplayed from '../../../helpers/assertions/VerifyNotificationDisplayed';
import { StartDriver, StopDriver } from '../../../helpers/context/WithDriver';

export default function() {
/**
 * When user provides bad password,
 * Then he gets a notification message.
 */
  describe('failed login as john in first factor', function() {
    this.timeout(10000);

    before(async function() {
      this.driver = await StartDriver();
      await VisitPageAndWaitUrlIs(this.driver, "https://login.example.com:8080/#/")
      await FillLoginPageWithUserAndPasswordAndClick(this.driver, 'john', 'bad_password');
    });

    after(async function() {
      await StopDriver(this.driver);
    })

    it('should get a notification message', async function () {
      await VerifyNotificationDisplayed(this.driver, "Authentication failed. Check your credentials.");
    });
  });
}

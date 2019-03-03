import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import VisitPage from "../../../helpers/VisitPage";
import VerifyIsAlreadyAuthenticatedStage from "../../../helpers/assertions/VerifyIsAlreadyAuthenticatedStage";
import RegisterAndLoginTwoFactor from "../../../helpers/behaviors/RegisterAndLoginTwoFactor";


export default function() {
  describe('When visiting /logout the user is redirected to already logged in page to log out', function() {
    before(async function() {
      this.driver = await StartDriver();
      await RegisterAndLoginTwoFactor(this.driver, 'john', "password", true);
    });

    after(async function() {
      await StopDriver(this.driver);
    });

    it('should redirect the user', async function() {
      await VisitPage(this.driver, 'https://login.example.com:8080/#/logout');
      await VerifyIsAlreadyAuthenticatedStage(this.driver);
    });
  });
}
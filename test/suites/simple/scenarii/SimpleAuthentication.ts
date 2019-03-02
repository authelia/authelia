import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import RegisterAndLoginTwoFactor from "../../../helpers/behaviors/RegisterAndLoginTwoFactor";
import VerifyUrlIs from "../../../helpers/assertions/VerifyUrlIs";
import VisitPage from "../../../helpers/VisitPage";
import VisitPageAndWaitUrlIs from "../../../helpers/behaviors/VisitPageAndWaitUrlIs";


export default function() {
  describe('The user is redirected to target url upon successful authentication', function() {
    before(async function() {
      this.driver = await StartDriver();
      await RegisterAndLoginTwoFactor(this.driver, 'john', "password", true, 'https://admin.example.com:8080/secret.html');
    });

    after(async function() {
      await StopDriver(this.driver);
    });

    it('should redirect the user', async function() {
      await VerifyUrlIs(this.driver, 'https://admin.example.com:8080/secret.html');
    });
  });

  describe('The target url is in "rd" parameter of the portal URL', function() {
    before(async function() {
      this.driver = await StartDriver();
      await VisitPage(this.driver, 'https://admin.example.com:8080/secret.html');
    });

    after(async function() {
      await StopDriver(this.driver);
    });

    it('should redirect the user', async function() {
      await VerifyUrlIs(this.driver, 'https://login.example.com:8080/?rd=https://admin.example.com:8080/secret.html');
    });
  })
}
import Logout from "../../../helpers/Logout";
import ChildProcess from 'child_process';
import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import VerifySecretObserved from "../../../helpers/assertions/VerifySecretObserved";
import RegisterAndLoginTwoFactor from "../../../helpers/behaviors/RegisterAndLoginTwoFactor";
import VisitPageAndWaitUrlIs from "../../../helpers/behaviors/VisitPageAndWaitUrlIs";
import { GET_Expect502 } from "../../../helpers/utils/Requests";
import LoginAndRegisterTotp from "../../../helpers/LoginAndRegisterTotp";
import FullLogin from "../../../helpers/FullLogin";

export default function() {
  describe('Session is still valid after Authelia restarts', function() {
    before(async function() {
      // Be sure to start fresh
      ChildProcess.execSync('rm -f .authelia-interrupt');

      this.driver = await StartDriver();
      await RegisterAndLoginTwoFactor(this.driver, 'john', "password", true, 'https://admin.example.com:8080/secret.html');
      await VisitPageAndWaitUrlIs(this.driver, 'https://home.example.com:8080/');
    });

    after(async function() {
      await Logout(this.driver);
      await StopDriver(this.driver);

      // Be sure to cleanup
      ChildProcess.execSync('rm -f .authelia-interrupt');
    });

    it("should still access the secret after Authelia restarted", async function() {
      ChildProcess.execSync('touch .authelia-interrupt');
      await GET_Expect502('https://login.example.com:8080/api/state');
      await this.driver.sleep(1000);
      ChildProcess.execSync('rm .authelia-interrupt');
      await this.driver.sleep(4000);

      await VisitPageAndWaitUrlIs(this.driver, 'https://admin.example.com:8080/secret.html');
      await VerifySecretObserved(this.driver);
    });  
  });

  describe('Secrets are persisted even if Authelia restarts', function() {
    before(async function() {
      // Be sure to start fresh
      ChildProcess.execSync('rm -f .authelia-interrupt');

      this.driver = await StartDriver();
      this.secret = await LoginAndRegisterTotp(this.driver, 'john', "password", true);
      await Logout(this.driver);
    });

    after(async function() {
      await Logout(this.driver);
      await StopDriver(this.driver);

      // Be sure to cleanup
      ChildProcess.execSync('rm -f .authelia-interrupt');
    });

    it("should still access the secret after Authelia restarted", async function() {
      ChildProcess.execSync('touch .authelia-interrupt');
      await GET_Expect502('https://login.example.com:8080/api/state');
      await this.driver.sleep(1000);
      ChildProcess.execSync('rm .authelia-interrupt');
      await this.driver.sleep(4000);

      // The user can re-authenticate with the secret.
      await FullLogin(this.driver, 'john', this.secret, 'https://admin.example.com:8080/secret.html')
    }); 
  });
}
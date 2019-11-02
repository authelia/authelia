import Logout from "../../../helpers/Logout";
import ChildProcess from 'child_process';
import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import VerifySecretObserved from "../../../helpers/assertions/VerifySecretObserved";
import VisitPageAndWaitUrlIs from "../../../helpers/behaviors/VisitPageAndWaitUrlIs";
import { GET_Expect502 } from "../../../helpers/utils/Requests";
import LoginAndRegisterTotp from "../../../helpers/LoginAndRegisterTotp";
import FullLogin from "../../../helpers/FullLogin";
import ValidateTotp from "../../../helpers/ValidateTotp";
import VerifyUrlIs from "../../../helpers/assertions/WaitUrlIs";

export default function() {
  describe('Session is still valid after Authelia restarts', function() {
    before(async function() {
      // Be sure to start fresh
      ChildProcess.execSync('rm -f .authelia-interrupt');

      this.driver = await StartDriver();
      this.secret = await LoginAndRegisterTotp(this.driver, 'john', "password", true);
      await VisitPageAndWaitUrlIs(this.driver, "https://login.example.com:8080/#/");
      await ValidateTotp(this.driver, this.secret);
      await VerifyUrlIs(this.driver, "https://home.example.com:8080/");

      ChildProcess.execSync('touch .authelia-interrupt');
      await GET_Expect502('https://login.example.com:8080/api/state');
      await this.driver.sleep(1000);
      ChildProcess.execSync('rm .authelia-interrupt');
      await this.driver.sleep(4000);
    });

    after(async function() {
      await Logout(this.driver);
      await StopDriver(this.driver);

      // Be sure to cleanup
      ChildProcess.execSync('rm -f .authelia-interrupt');
    });

    it("should still access the secret after Authelia restarted", async function() {
      await VisitPageAndWaitUrlIs(this.driver, "https://admin.example.com:8080/secret.html");
      await VerifySecretObserved(this.driver);
    });
    
    it("should still access the secret after Authelia restarted", async function() {
      await Logout(this.driver);
      // The user can re-authenticate with the secret.
      await FullLogin(this.driver, 'john', this.secret, "https://admin.example.com:8080/secret.html");
      await VerifyUrlIs(this.driver, "https://admin.example.com:8080/secret.html");
      await VerifySecretObserved(this.driver);
    }); 
  });
}
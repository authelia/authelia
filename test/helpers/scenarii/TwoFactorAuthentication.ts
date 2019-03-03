import { StartDriver, StopDriver } from "../context/WithDriver";
import RegisterAndLoginTwoFactor from "../behaviors/RegisterAndLoginTwoFactor";
import VerifyUrlIs from "../assertions/VerifyUrlIs";

export default function (timeout: number = 5000) {
  return function() {
    describe('The user is redirected to target url upon successful authentication', function() {
      before(async function() {
        this.driver = await StartDriver();
        await RegisterAndLoginTwoFactor(this.driver, 'john', "password", true, 'https://admin.example.com:8080/secret.html', timeout);
      });
  
      after(async function() {
        await StopDriver(this.driver);
      });
  
      it('should redirect the user', async function() {
        await VerifyUrlIs(this.driver, 'https://admin.example.com:8080/secret.html', timeout);
      });
    });
  }
}
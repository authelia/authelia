import Logout from "../../../helpers/Logout";
import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import RegisterAndLoginWith2FA from "../../../helpers/behaviors/RegisterAndLoginTwoFactor";
import VerifyForwardedHeaderIs from "../../../helpers/assertions/VerifyForwardedHeaderIs";
import LoginOneFactor from "../../../helpers/behaviors/LoginOneFactor";
import VerifyUrlIs from "../../../helpers/assertions/VerifyUrlIs";

export default function() {
  describe("Custom-Forwarded-User and Custom-Forwarded-Groups are correctly forwarded to protected backend", function() {
    this.timeout(100000);

    describe("Headers after single factor authentication", function() {
      before(async function() {
        this.driver = await StartDriver();
        await LoginOneFactor(this.driver, "john", "password", "https://singlefactor.example.com:8080/headers");
      });
      
      after(async function() {
        await Logout(this.driver);
        await StopDriver(this.driver);
      });

      it("should see header 'Custom-Forwarded-User' set to 'john'", async function() {
        await VerifyForwardedHeaderIs(this.driver, 'Custom-Forwarded-User', 'john');
      });
  
      it("should see header 'Custom-Forwarded-Groups' set to 'dev,admin'", async function() {
        await VerifyForwardedHeaderIs(this.driver, 'Custom-Forwarded-Groups', 'dev,admin');
      });
    });

    describe("Headers after two factor authentication", function() {
      before(async function() {
        this.driver = await StartDriver();
        await RegisterAndLoginWith2FA(this.driver, "john", "password", true, "https://secure.example.com:8080/headers");
      });
    
      after(async function() {
        await Logout(this.driver);
        await StopDriver(this.driver);
      });

      it("should see header 'Custom-Forwarded-User' set to 'john'", async function() {
        await VerifyForwardedHeaderIs(this.driver, 'Custom-Forwarded-User', 'john');
      });
  
      it("should see header 'Custom-Forwarded-Groups' set to 'dev,admin'", async function() {
        await VerifyForwardedHeaderIs(this.driver, 'Custom-Forwarded-Groups', 'dev,admin');
      });
    });
  });
}
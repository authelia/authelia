import Logout from "../../../helpers/Logout";
import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import VerifyForwardedHeaderIs from "../../../helpers/assertions/VerifyForwardedHeaderIs";
import LoginOneFactor from "../../../helpers/behaviors/LoginOneFactor";
import VisitPageAndWaitUrlIs from "../../../helpers/behaviors/VisitPageAndWaitUrlIs";
import VerifyButtonDoesNotExist from "../../../helpers/assertions/VerifyButtonDoesNotExist";

export default function() {
  describe("Custom-Forwarded-User and Custom-Forwarded-Groups are correctly forwarded when available", function() {
    this.timeout(100000);

    describe("Headers are not forwarded for anonymous user", function() {
      before(async function() {
        this.driver = await StartDriver();
        await VisitPageAndWaitUrlIs(this.driver, "https://public.example.com:8080/headers");
      });
      
      after(async function() {
        await Logout(this.driver);
        await StopDriver(this.driver);
      });

      it("should check header 'Custom-Forwarded-User' does not exist", async function() {
        await VerifyButtonDoesNotExist(this.driver, 'Custom-Forwarded-User');
      });
  
      it("should check header 'Custom-Forwarded-Groups' does not exist", async function() {
        await VerifyButtonDoesNotExist(this.driver, 'Custom-Forwarded-Groups');
      });
    });

    describe("Headers are forwarded for authenticated user", function() {
      before(async function() {
        this.driver = await StartDriver();
        await LoginOneFactor(this.driver, "john", "password", "https://public.example.com:8080/headers");
      });
      
      after(async function() {
        await Logout(this.driver);
        await StopDriver(this.driver);
      });

      it("should see header 'Custom-Forwarded-User' set to 'john'", async function() {
        await VerifyForwardedHeaderIs(this.driver, 'Custom-Forwarded-User', 'john');
      });
  
      it("should see header 'Custom-Forwarded-Groups' set to 'admins,dev'", async function() {
        await VerifyForwardedHeaderIs(this.driver, 'Custom-Forwarded-Groups', 'admins,dev');
      });
    });
  });
}
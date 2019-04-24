import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import LoginAndRegisterTotp from "../../../helpers/LoginAndRegisterTotp";
import FillLoginPageAndClick from "../../../helpers/FillLoginPageAndClick";
import ValidateTotp from "../../../helpers/ValidateTotp";
import VerifySecretObserved from "../../../helpers/assertions/VerifySecretObserved";
import VisitPageAndWaitUrlIs from "../../../helpers/behaviors/VisitPageAndWaitUrlIs";
import VerifyUrlIs from "../../../helpers/assertions/VerifyUrlIs";
import VisitPage from "../../../helpers/VisitPage";

async function createClient(id: number) {
  return await StartDriver({
    proxyType: "manual",
    httpProxy: `http://proxy-client${id}.example.com:3128`
  });
}

export default function() {
  before(async function() {
    const driver = await StartDriver();
    this.secret = await LoginAndRegisterTotp(driver, "john", "password", true);
    if (!this.secret) throw new Error('No secret!');
    await StopDriver(driver);
  });

  describe("Standard client (from public network)", function() {
    before(async function() {
      this.driver = await StartDriver();
    });

    after(async function() {
      await StopDriver(this.driver);
    });

    it("should require two factor", async function() {
      await VisitPage(this.driver, "https://secure.example.com:8080/secret.html");
      await VerifyUrlIs(this.driver, "https://login.example.com:8080/#/?rd=https://secure.example.com:8080/secret.html");
      await FillLoginPageAndClick(this.driver, "john", "password");
      await ValidateTotp(this.driver, this.secret);
      await VerifyUrlIs(this.driver, "https://secure.example.com:8080/secret.html");
      await VerifySecretObserved(this.driver);
    });
  })

  describe("Client 1 (from network 192.168.240.201/32)", function() {
    before(async function() {
      this.client1 = await createClient(1);
    });

    after(async function() {
      await StopDriver(this.client1);
    });

    it("should require one factor", async function() {
      await VisitPage(this.client1, "https://secure.example.com:8080/secret.html");
      await VerifyUrlIs(this.client1, "https://login.example.com:8080/#/?rd=https://secure.example.com:8080/secret.html");
      await FillLoginPageAndClick(this.client1, 'john', 'password');
      await VerifyUrlIs(this.client1, "https://secure.example.com:8080/secret.html");
      await VerifySecretObserved(this.client1);
    });
  });

  describe("Client 2  (from network 192.168.240.202/32)", function() {
    before(async function() {
      this.client2 = await createClient(2);
    });

    after(async function() {
      await StopDriver(this.client2);
    });

    it("should bypass", async function() {
      await VisitPageAndWaitUrlIs(this.client2, "https://secure.example.com:8080/secret.html");
      await VerifyUrlIs(this.client2, "https://secure.example.com:8080/secret.html");
      await VerifySecretObserved(this.client2);
    });
  });
}
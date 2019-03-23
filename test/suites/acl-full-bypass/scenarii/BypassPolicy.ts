import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import VerifySecretObserved from "../../../helpers/assertions/VerifySecretObserved";
import VisitPageAndWaitUrlIs from "../../../helpers/behaviors/VisitPageAndWaitUrlIs";

export default function() {
  before(async function() {
    this.driver = await StartDriver();
  });

  after(async function () {
    await StopDriver(this.driver);
  });

  it('should have access to admin.example.com/secret.html', async function () {
    await VisitPageAndWaitUrlIs(this.driver, "https://admin.example.com:8080/secret.html");
    await VerifySecretObserved(this.driver);
  });

  it('should have access to public.example.com/secret.html', async function () {
    await VisitPageAndWaitUrlIs(this.driver, "https://public.example.com:8080/secret.html");
    await VerifySecretObserved(this.driver);
  });
}
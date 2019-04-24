import LoginAndRegisterTotp from "../../../helpers/LoginAndRegisterTotp";
import VerifySecretObserved from "../../../helpers/assertions/VerifySecretObserved";
import { StartDriver, StopDriver } from "../../../helpers/context/WithDriver";
import ValidateTotp from "../../../helpers/ValidateTotp";
import Logout from "../../../helpers/Logout";
import VisitPageAndWaitUrlIs from "../../../helpers/behaviors/VisitPageAndWaitUrlIs";
import VerifyBodyContains from "../../../helpers/assertions/VerifyBodyContains";
import VerifyUrlIs from "../../../helpers/assertions/WaitUrlIs";

async function ShouldHaveAccessTo(url: string) {
  it('should have access to ' + url, async function() {
    await VisitPageAndWaitUrlIs(this.driver, url);
    await VerifySecretObserved(this.driver);
  })
}

async function ShouldNotHaveAccessTo(url: string) {
  it('should not have access to ' + url, async function() {
    await VisitPageAndWaitUrlIs(this.driver, url);
    await VerifyBodyContains(this.driver, "403 Forbidden");
  })
}

// we verify that the user has only access to want he is granted to.
export default function() {
  // We ensure that bob has access to what he is granted to
  describe('Permissions of user john', function() {
    before(async function() {
      this.driver = await StartDriver();
      const secret = await LoginAndRegisterTotp(this.driver, "john", "password", true);
      await VisitPageAndWaitUrlIs(this.driver, 'https://login.example.com:8080/#/');
      await ValidateTotp(this.driver, secret);
      await VerifyUrlIs(this.driver, "https://home.example.com:8080/");
    });

    after(async function() {
      await Logout(this.driver);
      await StopDriver(this.driver);
    });

    ShouldHaveAccessTo('https://public.example.com:8080/secret.html');
    ShouldHaveAccessTo('https://secure.example.com:8080/secret.html');
    ShouldHaveAccessTo('https://dev.example.com:8080/groups/admin/secret.html');
    ShouldHaveAccessTo('https://dev.example.com:8080/groups/dev/secret.html');
    ShouldHaveAccessTo('https://dev.example.com:8080/users/john/secret.html');
    ShouldHaveAccessTo('https://dev.example.com:8080/users/harry/secret.html');
    ShouldHaveAccessTo('https://dev.example.com:8080/users/bob/secret.html');
    ShouldHaveAccessTo('https://admin.example.com:8080/secret.html');
    ShouldHaveAccessTo('https://mx1.mail.example.com:8080/secret.html');
    ShouldHaveAccessTo('https://singlefactor.example.com:8080/secret.html');
    ShouldNotHaveAccessTo('https://mx2.mail.example.com:8080/secret.html');
  })

  // We ensure that bob has access to what he is granted to
  describe('Permissions of user bob', function() {
    before(async function() {
      this.driver = await StartDriver();
      const secret = await LoginAndRegisterTotp(this.driver, "bob", "password", true);
      await VisitPageAndWaitUrlIs(this.driver, 'https://login.example.com:8080/#/');
      await ValidateTotp(this.driver, secret);
      await VerifyUrlIs(this.driver, "https://home.example.com:8080/");
    });

    after(async function() {
      await Logout(this.driver);
      await StopDriver(this.driver);
    });

    ShouldHaveAccessTo('https://public.example.com:8080/secret.html');
    ShouldHaveAccessTo('https://secure.example.com:8080/secret.html');
    ShouldNotHaveAccessTo('https://dev.example.com:8080/groups/admin/secret.html');
    ShouldHaveAccessTo('https://dev.example.com:8080/groups/dev/secret.html');
    ShouldNotHaveAccessTo('https://dev.example.com:8080/users/john/secret.html');
    ShouldNotHaveAccessTo('https://dev.example.com:8080/users/harry/secret.html');
    ShouldHaveAccessTo('https://dev.example.com:8080/users/bob/secret.html');
    ShouldNotHaveAccessTo('https://admin.example.com:8080/secret.html');
    ShouldHaveAccessTo('https://mx1.mail.example.com:8080/secret.html');
    ShouldHaveAccessTo('https://singlefactor.example.com:8080/secret.html');
    ShouldHaveAccessTo('https://mx2.mail.example.com:8080/secret.html');
  });

  // We ensure that harry has access to what he is granted to
  describe('Permissions of user harry', function() {
    before(async function() {
      this.driver = await StartDriver();
      const secret = await LoginAndRegisterTotp(this.driver, "harry", "password", true);
      await VisitPageAndWaitUrlIs(this.driver, 'https://login.example.com:8080/#/');
      await ValidateTotp(this.driver, secret);
      await VerifyUrlIs(this.driver, "https://home.example.com:8080/");
    });

    after(async function() {
      await Logout(this.driver);
      await StopDriver(this.driver);
    });

    ShouldHaveAccessTo('https://public.example.com:8080/secret.html');
    ShouldHaveAccessTo('https://secure.example.com:8080/secret.html');
    ShouldNotHaveAccessTo('https://dev.example.com:8080/groups/admin/secret.html');
    ShouldNotHaveAccessTo('https://dev.example.com:8080/groups/dev/secret.html');
    ShouldNotHaveAccessTo('https://dev.example.com:8080/users/john/secret.html');
    ShouldHaveAccessTo('https://dev.example.com:8080/users/harry/secret.html');
    ShouldNotHaveAccessTo('https://dev.example.com:8080/users/bob/secret.html');
    ShouldNotHaveAccessTo('https://admin.example.com:8080/secret.html');
    ShouldNotHaveAccessTo('https://mx1.mail.example.com:8080/secret.html');
    ShouldHaveAccessTo('https://singlefactor.example.com:8080/secret.html');
    ShouldNotHaveAccessTo('https://mx2.mail.example.com:8080/secret.html');
  });
}
require("chromedriver");
import seleniumWebdriver = require("selenium-webdriver");
import Cucumber = require("cucumber");
import Fs = require("fs");
import Speakeasy = require("speakeasy");
import Assert = require("assert");

function CustomWorld() {
  const that = this;
  this.driver = new seleniumWebdriver.Builder()
    .forBrowser("chrome")
    .build();

  this.totpSecrets = {};
  this.configuration = {};

  this.visit = function (link: string) {
    return this.driver.get(link);
  };

  this.setFieldTo = function (fieldName: string, content: string) {
    return this.driver.findElement(seleniumWebdriver.By.id(fieldName))
      .sendKeys(content);
  };

  this.clearField = function (fieldName: string) {
    return this.driver.findElement(seleniumWebdriver.By.id(fieldName)).clear();
  };

  this.getErrorPage = function (code: number) {
    const that = this;
    return this.driver.wait(seleniumWebdriver.until.elementLocated(seleniumWebdriver.By.tagName("h1")), 2000)
      .then(function () {
        return that.driver
          .findElement(seleniumWebdriver.By.tagName("h1")).getText();
      })
      .then(function (txt: string) {
        Assert.equal(txt, "Error " + code);
      });
  };

  this.clickOnButton = function (buttonText: string) {
    const that = this;
    return this.driver.wait(seleniumWebdriver.until.elementLocated(seleniumWebdriver.By.tagName("button")), 2000)
      .then(function () {
        return that.driver
          .findElement(seleniumWebdriver.By.tagName("button"))
          .findElement(seleniumWebdriver.By.xpath("//button[contains(.,'" + buttonText + "')]"))
          .click();
      });
  };

  this.loginWithUserPassword = function (username: string, password: string) {
    return that.driver.wait(seleniumWebdriver.until.elementLocated(seleniumWebdriver.By.id("username")), 4000)
      .then(function () {
        return that.driver.findElement(seleniumWebdriver.By.id("username"))
          .sendKeys(username);
      })
      .then(function () {
        return that.driver.findElement(seleniumWebdriver.By.id("password"))
          .sendKeys(password);
      })
      .then(function () {
        return that.driver.findElement(seleniumWebdriver.By.tagName("button"))
          .click();
      });
  };

  this.registerTotpSecret = function (totpSecretHandle: string) {
    return that.driver.wait(seleniumWebdriver.until.elementLocated(seleniumWebdriver.By.className("register-totp")), 4000)
      .then(function () {
        return that.driver.findElement(seleniumWebdriver.By.className("register-totp")).click();
      })
      .then(function () {
        const notif = Fs.readFileSync("/tmp/notifications/notification.txt").toString();
        const regexp = new RegExp(/Link: (.+)/);
        const match = regexp.exec(notif);
        const link = match[1];
        console.log("Link: " + link);
        return that.driver.get(link);
      })
      .then(function () {
        return that.driver.wait(seleniumWebdriver.until.elementLocated(seleniumWebdriver.By.id("secret")), 5000);
      })
      .then(function () {
        return that.driver.findElement(seleniumWebdriver.By.id("secret")).getText();
      })
      .then(function (secret: string) {
        that.totpSecrets[totpSecretHandle] = secret;
      });
  };

  this.useTotpTokenHandle = function (totpSecretHandle: string) {
    if (!this.totpSecrets[totpSecretHandle])
      throw new Error("No available TOTP token handle " + totpSecretHandle);

    const token = Speakeasy.totp({
      secret: this.totpSecrets[totpSecretHandle],
      encoding: "base32"
    });
    return this.useTotpToken(token);
  };

  this.useTotpToken = function (totpSecret: string) {
    return that.driver.wait(seleniumWebdriver.until.elementLocated(seleniumWebdriver.By.id("token")), 5000)
      .then(function () {
        return that.driver.findElement(seleniumWebdriver.By.id("token"))
          .sendKeys(totpSecret);
      });
  };

  this.registerTotpAndSignin = function (username: string, password: string) {
    const totpHandle = "HANDLE";
    const authUrl = "https://auth.test.local:8080/";
    const that = this;
    return this.visit(authUrl)
      .then(function () {
        return that.loginWithUserPassword(username, password);
      })
      .then(function () {
        return that.registerTotpSecret(totpHandle);
      })
      .then(function () {
        return that.visit(authUrl);
      })
      .then(function () {
        return that.loginWithUserPassword(username, password);
      })
      .then(function () {
        return that.useTotpTokenHandle(totpHandle);
      })
      .then(function () {
        return that.clickOnButton("TOTP");
      });
  };
}

Cucumber.defineSupportCode(function ({ setWorldConstructor }) {
  setWorldConstructor(CustomWorld);
});
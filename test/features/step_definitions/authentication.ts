import Cucumber = require("cucumber");
import seleniumWebdriver = require("selenium-webdriver");
import Assert = require("assert");
import Fs = require("fs");
import Speakeasy = require("speakeasy");
import CustomWorld = require("../support/world");

Cucumber.defineSupportCode(function ({ Given, When, Then }) {
  When(/^I visit "(https:\/\/[a-z0-9:.\/=?-]+)"$/, function (link: string) {
    return this.visit(link);
  });

  When("I set field {stringInDoubleQuotes} to {stringInDoubleQuotes}", function (fieldName: string, content: string) {
    return this.setFieldTo(fieldName, content);
  });

  When("I click on {stringInDoubleQuotes}", function (text: string) {
    return this.clickOnButton(text);
  });

  Given("I login with user {stringInDoubleQuotes} and password {stringInDoubleQuotes}", function (username: string, password: string) {
    return this.loginWithUserPassword(username, password);
  });

  Given("I register a TOTP secret called {stringInDoubleQuotes}", function (handle: string) {
    return this.registerTotpSecret(handle);
  });

  Given("I use {stringInDoubleQuotes} as TOTP token", function (token: string) {
    return this.useTotpToken(token);
  });

  Given("I use {stringInDoubleQuotes} as TOTP token handle", function (handle) {
    return this.useTotpTokenHandle(handle);
  });

  Then("I get a notification with message {stringInDoubleQuotes}", function (notificationMessage: string) {
    const that = this;
    that.driver.sleep(500);
    return this.driver
      .findElement(seleniumWebdriver.By.className("notifyjs-corner"))
      .findElement(seleniumWebdriver.By.tagName("span"))
      .findElement(seleniumWebdriver.By.xpath("//span[contains(.,'" + notificationMessage + "')]"));
  });

  When("I visit {stringInDoubleQuotes} and get redirected {stringInDoubleQuotes}", function (url: string, redirectUrl: string) {
    const that = this;
    return this.driver.get(url)
      .then(function () {
        return that.driver.wait(seleniumWebdriver.until.urlIs(redirectUrl), 2000);
      });
  });

  Given("I register TOTP and login with user {stringInDoubleQuotes} and password {stringInDoubleQuotes}", function (username: string, password: string) {
    return this.registerTotpAndSignin(username, password);
  });

  function hasAccessToSecret(link: string, driver: any) {
    return driver.get(link)
      .then(function () {
        return driver.findElement(seleniumWebdriver.By.tagName("body")).getText()
          .then(function (body: string) {
            Assert(body.indexOf("This is a very important secret!") > -1);
          });
      });
  }

  function hasNoAccessToSecret(link: string, driver: any) {
    return driver.get(link)
      .then(function () {
        return driver.wait(seleniumWebdriver.until.urlIs("https://auth.test.local:8080/"));
      });
  }

  Then("I have access to:", function (dataTable: Cucumber.TableDefinition) {
    const promises = [];
    for (let i = 0; i < dataTable.rows().length; i++) {
      const url = (dataTable.hashes() as any)[i].url;
      promises.push(hasAccessToSecret(url, this.driver));
    }
    return Promise.all(promises);
  });

  Then("I have no access to:", function (dataTable: Cucumber.TableDefinition) {
    const promises = [];
    for (let i = 0; i < dataTable.rows().length; i++) {
      const url = (dataTable.hashes() as any)[i].url;
      promises.push(hasNoAccessToSecret(url, this.driver));
    }
    return Promise.all(promises);
  });
});
import Cucumber = require("cucumber");
import seleniumWebdriver = require("selenium-webdriver");
import Assert = require("assert");
import Fs = require("fs");
import Speakeasy = require("speakeasy");
import CustomWorld = require("../support/world");
import BluebirdPromise = require("bluebird");

Cucumber.defineSupportCode(function ({ Given, When, Then }) {
  When(/^I visit "(https:\/\/[a-zA-Z0-9:%&._\/=?-]+)"$/, function (link: string) {
    return this.visit(link);
  });

  When("I set field {stringInDoubleQuotes} to {stringInDoubleQuotes}", function (fieldName: string, content: string) {
    return this.setFieldTo(fieldName, content);
  });

  When("I clear field {stringInDoubleQuotes}", function (fieldName: string) {
    return this.clearField(fieldName);
  });

  When("I click on {stringInDoubleQuotes}", function (text: string) {
    return this.clickOnButton(text);
  });

  Given("I login with user {stringInDoubleQuotes} and password {stringInDoubleQuotes}",
    function (username: string, password: string) {
      return this.loginWithUserPassword(username, password);
    });

  Given("I login with user {stringInDoubleQuotes} and password {stringInDoubleQuotes} \
and I use TOTP token handle {stringInDoubleQuotes}",
    function (username: string, password: string, totpTokenHandle: string) {
      const that = this;
      return this.loginWithUserPassword(username, password)
        .then(function () {
          return that.useTotpTokenHandle(totpTokenHandle);
        });
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

  When("I visit {stringInDoubleQuotes} and get redirected {stringInDoubleQuotes}",
    function (url: string, redirectUrl: string) {
      const that = this;
      return this.driver.get(url)
        .then(function () {
          return that.driver.wait(seleniumWebdriver.until.urlIs(redirectUrl), 2000);
        });
    });

  Given("I register TOTP and login with user {stringInDoubleQuotes} and password {stringInDoubleQuotes}",
    function (username: string, password: string) {
      return this.registerTotpAndSignin(username, password);
    });

  function hasAccessToSecret(link: string, that: any) {
    return that.driver.get(link)
      .then(function () {
        return that.waitUntilUrlContains(link);
      });
  }

  function hasNoAccessToSecret(link: string, that: any) {
    return that.driver.get(link)
      .then(function () {
        return that.getErrorPage(403);
      });
  }

  Then("I have access to:", function (dataTable: Cucumber.TableDefinition) {
    const promises: any = [];
    for (let i = 0; i < dataTable.rows().length; i++) {
      const url = (dataTable.hashes() as any)[i].url;
      promises.push(hasAccessToSecret(url, this));
    }
    return BluebirdPromise.all(promises);
  });

  Then("I have no access to:", function (dataTable: Cucumber.TableDefinition) {
    const promises = [];
    for (let i = 0; i < dataTable.rows().length; i++) {
      const url = (dataTable.hashes() as any)[i].url;
      promises.push(hasNoAccessToSecret(url, this));
    }
    return BluebirdPromise.all(promises);
  });
});
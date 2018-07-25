import {Given, When, Then, TableDefinition} from "cucumber";
import seleniumWebdriver = require("selenium-webdriver");
import Assert = require("assert");
import Fs = require("fs");
import Speakeasy = require("speakeasy");
import CustomWorld = require("../support/world");
import BluebirdPromise = require("bluebird");
import Request = require("request-promise");

When(/^I visit "(https:\/\/[a-zA-Z0-9:%&._\/=?-]+)"$/, function (link: string) {
  return this.visit(link);
});

When("I wait for notification to disappear", function () {
  const that = this;
  const notificationEl = this.driver.findElement(seleniumWebdriver.By.className("notification"));
  return this.driver.wait(seleniumWebdriver.until.elementIsVisible(notificationEl), 15000)
    .then(function () {
      return that.driver.wait(seleniumWebdriver.until.elementIsNotVisible(notificationEl), 15000);
    })
})

When("I set field {string} to {string}", function (fieldName: string, content: string) {
  return this.setFieldTo(fieldName, content);
});

When("I clear field {string}", function (fieldName: string) {
  return this.clearField(fieldName);
});

When("I click on {string}", function (text: string) {
  return this.clickOnButton(text);
});

Given("I login with user {string} and password {string}",
  function (username: string, password: string) {
    return this.loginWithUserPassword(username, password);
  });

Given("I login with user {string} and password {string} \
and I use TOTP token handle {string}",
  function (username: string, password: string, totpTokenHandle: string) {
    const that = this;
    return this.loginWithUserPassword(username, password)
      .then(function () {
        return that.useTotpTokenHandle(totpTokenHandle);
      });
  });

Given("I register a TOTP secret called {string}", function (handle: string) {
  return this.registerTotpSecret(handle);
});

Given("I use {string} as TOTP token", function (token: string) {
  return this.useTotpToken(token);
});

Given("I use {string} as TOTP token handle", function (handle) {
  return this.useTotpTokenHandle(handle);
});

When("I visit {string} and get redirected {string}",
  function (url: string, redirectUrl: string) {
    const that = this;
    return this.driver.get(url)
      .then(function () {
        return that.driver.wait(seleniumWebdriver.until.urlIs(redirectUrl), 6000);
      });
  });

Given("I register TOTP and login with user {string} and password {string}",
  function (username: string, password: string) {
    return this.registerTotpAndSignin(username, password);
  });

function endpointReplyWith(context: any, link: string, method: string,
  returnCode: number) {
  return Request(link, {
    method: method
  })
    .then(function (response: string) {
      Assert(response.indexOf("Error " + returnCode) >= 0);
      return BluebirdPromise.resolve();
    }, function (response: any) {
      Assert.equal(response.statusCode, returnCode);
      return BluebirdPromise.resolve();
    });
}

Then("the following endpoints reply with:", function (dataTable: TableDefinition) {
  const promises = [];
  for (let i = 0; i < dataTable.rows().length; i++) {
    const url: string = (dataTable.hashes() as any)[i].url;
    const method: string = (dataTable.hashes() as any)[i].method;
    const code: number = (dataTable.hashes() as any)[i].code;
    promises.push(endpointReplyWith(this, url, method, code));
  }
  return BluebirdPromise.all(promises);
});
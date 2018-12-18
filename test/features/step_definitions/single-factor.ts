import {When, Then} from "cucumber";
import seleniumWebdriver = require("selenium-webdriver");
import Request = require("request-promise");
import BluebirdPromise = require("bluebird");
import Util = require("util");

When("I request {string} with username {string}" +
  " and password {string} using basic authentication",
  function (url: string, username: string, password: string) {
    const that = this;
    return Request(url, {
      auth: {
        username: username,
        password: password
      },
      resolveWithFullResponse: true
    })
      .then(function (response: any) {
        that.response = response;
      });
  });

Then("I receive the secret page", function () {
  if (this.response.body.match("This is a very important secret!"))
    return BluebirdPromise.resolve();
  return BluebirdPromise.reject(new Error("Secret page not received."));
});

Then("I received header {string} set to {string}",
  function (expectedHeaderName: string, expectedValue: string) {
    const expectedLine = Util.format("\"%s\": \"%s\"", expectedHeaderName,
      expectedValue);
    if (this.response.body.indexOf(expectedLine) > 0)
      return BluebirdPromise.resolve();
    return BluebirdPromise.reject(new Error(
      Util.format("No such header or with unexpected value.")));
  });

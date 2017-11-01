import Cucumber = require("cucumber");
import seleniumWebdriver = require("selenium-webdriver");
import Request = require("request-promise");
import BluebirdPromise = require("bluebird");
import Util = require("util");

Cucumber.defineSupportCode(function ({ Given, When, Then }) {
  When("I request {stringInDoubleQuotes} with username {stringInDoubleQuotes}" +
    " and password {stringInDoubleQuotes} using basic authentication",
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

  Then("I received header {stringInDoubleQuotes} set to {stringInDoubleQuotes}",
    function (expectedHeaderName: string, expectedValue: string) {
      const expectedLine = Util.format("\"%s\": \"%s\"", expectedHeaderName,
        expectedValue);
      if (this.response.body.indexOf(expectedLine) > 0)
        return BluebirdPromise.resolve();
      return BluebirdPromise.reject(new Error(
        Util.format("No such header or with unexpected value.")));
    })
});
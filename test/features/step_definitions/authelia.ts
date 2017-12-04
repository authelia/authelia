import Cucumber = require("cucumber");
import seleniumWebdriver = require("selenium-webdriver");
import Assert = require("assert");
import Request = require("request-promise");
import Bluebird = require("bluebird");

Cucumber.defineSupportCode(function ({ Given, When, Then, Before, After }) {
  Before(function () {
    this.jar = Request.jar();
  })

  When("I query {stringInDoubleQuotes}", function (url: string) {
    const that = this;
    return Request(url, { followRedirect: false })
      .then(function(response) {
        console.log(response);
        that.response = response;
      })
      .catch(function(err: Error) {
        that.error = err;
      })
  });

  Then("I get error code 401", function() {
    const that = this;
    return new Bluebird(function(resolve, reject) {
      if(that.error && that.error.statusCode == 401) {
        resolve();
      }
      else {
        if(that.response) 
          reject(new Error("No error thrown"));
        else if(that.error.statusCode != 401)
          reject(new Error("Error code != 401"));
      }
    });
  });

  Then("I get redirected to {stringInDoubleQuotes}", function(url: string) {
    const that = this;
    return new Bluebird(function(resolve, reject) {
      if(that.error && that.error.statusCode == 302 
        && that.error.message.indexOf(url) > -1) {
        resolve();
      }
      else {
        reject(new Error("Not redirected"));
      }
    });
  })
});
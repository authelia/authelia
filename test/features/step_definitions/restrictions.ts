import {Before, When, Then, TableDefinition} from "cucumber";
import seleniumWebdriver = require("selenium-webdriver");
import Assert = require("assert");
import Request = require("request-promise");
import Bluebird = require("bluebird");

Before(function () {
  this.jar = Request.jar();
});

Then("I get an error {int}", function (code: number) {
  return this.getErrorPage(code);
});

When("I request {string} with method {string}",
  function (url: string, method: string) {
    const that = this;
  });

function requestAndExpectStatusCode(ctx: any, url: string, method: string,
  expectedStatusCode: number) {
  return Request(url, {
    method: method,
    jar: ctx.jar
  })
    .then(function (body: string) {
      return Bluebird.resolve(parseInt(body.match(/Error ([0-9]{3})/)[1]));
    }, function (response: any) {
      return Bluebird.resolve(response.statusCode)
    })
    .then(function (statusCode: number) {
      try {
        Assert.equal(statusCode, expectedStatusCode);
      }
      catch (e) {
        console.log("%s (actual) != %s (expected)", statusCode,
          expectedStatusCode);
        throw e;
      }
    })
}

Then("I get the following status code when requesting:",
  function (dataTable: TableDefinition) {
    const promises: Bluebird<void>[] = [];
    for (let i = 0; i < dataTable.rows().length; i++) {
      const url: string = (dataTable.hashes() as any)[i].url;
      const method: string = (dataTable.hashes() as any)[i].method;
      const code: number = (dataTable.hashes() as any)[i].code;
      promises.push(requestAndExpectStatusCode(this, url, method, code));
    }
    return Bluebird.all(promises);
  })

When("I post {string} with body:", function (url: string,
  dataTable: TableDefinition) {
  const body = {};
  for (let i = 0; i < dataTable.rows().length; i++) {
    const key = (dataTable.hashes() as any)[i].key;
    const value = (dataTable.hashes() as any)[i].value;
    body[key] = value;
  }
  return Request.post(url, {
    body: body,
    jar: this.jar,
    json: true
  });
});
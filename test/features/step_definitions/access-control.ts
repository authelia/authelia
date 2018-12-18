import {Then} from "cucumber";

Then("I have access to {string}", function(url: string) {
  const that = this;
  return this.driver.get(url)
    .then(function () {
      return that.waitUntilUrlContains(url);
    });
});

Then("I have no access to {string}", function(url: string) {
  const that = this;
  return this.driver.get(url)
    .then(function () {
      return that.getErrorPage(403);
    });
});
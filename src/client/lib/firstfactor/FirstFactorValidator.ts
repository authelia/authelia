
import BluebirdPromise = require("bluebird");
import Endpoints = require("../../../server/endpoints");
import Constants = require("../../../server/constants");

export function validate(username: string, password: string,
  redirectUrl: string, onlyBasicAuth: boolean, $: JQueryStatic): BluebirdPromise<string> {
  return new BluebirdPromise<string>(function (resolve, reject) {
    const url = Endpoints.FIRST_FACTOR_POST + "?" + Constants.REDIRECT_QUERY_PARAM + "=" + redirectUrl
      + "&" + Constants.ONLY_BASIC_AUTH_QUERY_PARAM + "=" + onlyBasicAuth;

    $.ajax({
      method: "POST",
      url: url,
      data: {
        username: username,
        password: password,
      }
    })
      .done(function (data: { redirect: string }) {
        resolve(data.redirect);
      })
      .fail(function (xhr: JQueryXHR, textStatus: string) {
        reject(new Error("Authetication failed. Please check your credentials."));
      });
  });
}

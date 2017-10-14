
import BluebirdPromise = require("bluebird");
import Endpoints = require("../../../../shared/api");
import Constants = require("../../../../shared/constants");
import Util = require("util");
import UserMessages = require("../../../../shared/UserMessages");

export function validate(username: string, password: string,
  redirectUrl: string, $: JQueryStatic): BluebirdPromise<string> {
  return new BluebirdPromise<string>(function (resolve, reject) {
    let url: string;
    if (redirectUrl != undefined) {
      const redirectParam = Util.format("%s=%s", Constants.REDIRECT_QUERY_PARAM, redirectUrl);
      url = Util.format("%s?%s", Endpoints.FIRST_FACTOR_POST, redirectParam);
    }
    else {
      url = Util.format("%s", Endpoints.FIRST_FACTOR_POST);
    }

    $.ajax({
      method: "POST",
      url: url,
      data: {
        username: username,
        password: password,
      }
    })
      .done(function (body: any) {
        if (body && body.error) {
          reject(new Error(body.error));
          return;
        }
        resolve(body.redirect);
      })
      .fail(function (xhr: JQueryXHR, textStatus: string) {
        reject(new Error(UserMessages.AUTHENTICATION_FAILED));
      });
  });
}

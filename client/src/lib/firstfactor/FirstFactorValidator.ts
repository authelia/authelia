
import BluebirdPromise = require("bluebird");
import Endpoints = require("../../../../shared/api");
import Constants = require("../../../../shared/constants");
import Util = require("util");
import UserMessages = require("../../../../shared/UserMessages");

export function validate(username: string, password: string,
  keepMeLoggedIn: boolean, redirectUrl: string, $: JQueryStatic)
  : BluebirdPromise<string> {
  return new BluebirdPromise<string>(function (resolve, reject) {
    let url: string;
    if (redirectUrl != undefined) {
      const redirectParam = Util.format("%s=%s", Constants.REDIRECT_QUERY_PARAM, redirectUrl);
      url = Util.format("%s?%s", Endpoints.FIRST_FACTOR_POST, redirectParam);
    }
    else {
      url = Util.format("%s", Endpoints.FIRST_FACTOR_POST);
    }

    const data: any = {
      username: username,
      password: password,
    };

    if (keepMeLoggedIn) {
      data.keepMeLoggedIn = "true";
    }

    $.ajax({
      method: "POST",
      url: url,
      data: data
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

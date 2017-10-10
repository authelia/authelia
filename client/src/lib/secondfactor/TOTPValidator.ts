
import BluebirdPromise = require("bluebird");
import Endpoints = require("../../../../shared/api");

export function validate(token: string, $: JQueryStatic): BluebirdPromise<string> {
  return new BluebirdPromise<string>(function (resolve, reject) {
    $.ajax({
      url: Endpoints.SECOND_FACTOR_TOTP_POST,
      data: {
        token: token,
      },
      method: "POST",
      dataType: "json"
    } as JQueryAjaxSettings)
      .done(function (body: any) {
        if (body && body.error) {
          reject(new Error(body.error));
          return;
        }
        resolve(body);
      })
      .fail(function (xhr: JQueryXHR, textStatus: string) {
        reject(new Error(textStatus));
      });
  });
}
import U2f = require("u2f");
import U2fApi from "u2f-api";
import BluebirdPromise = require("bluebird");
import Endpoints = require("../../../../shared/api");
import UserMessages = require("../../../../shared/UserMessages");
import { INotifier } from "../INotifier";
import { RedirectionMessage } from "../../../../shared/RedirectionMessage";
import { ErrorMessage } from "../../../../shared/ErrorMessage";
import GetPromised from "../GetPromised";

function finishU2fAuthentication(responseData: U2fApi.SignResponse,
  $: JQueryStatic): BluebirdPromise<string> {
  return new BluebirdPromise<string>(function (resolve, reject) {
    $.ajax({
      url: Endpoints.SECOND_FACTOR_U2F_SIGN_POST,
      data: responseData,
      method: "POST",
      dataType: "json"
    } as JQueryAjaxSettings)
      .done(function (body: RedirectionMessage | ErrorMessage) {
        if (body && "error" in body) {
          reject(new Error((body as ErrorMessage).error));
          return;
        }
        resolve((body as RedirectionMessage).redirect);
      })
      .fail(function (xhr: JQueryXHR, textStatus: string) {
        reject(new Error(textStatus));
      });
  });
}

export function validate($: JQueryStatic): BluebirdPromise<string> {
  return GetPromised($, Endpoints.SECOND_FACTOR_U2F_SIGN_REQUEST_GET, {},
    undefined, "json")
    .then(function (signRequest: U2f.Request) {
      return U2fApi.sign(signRequest, 60);
    })
    .then(function (signResponse: U2fApi.SignResponse) {
      return finishU2fAuthentication(signResponse, $);
    });
}

import U2f = require("u2f");
import U2fApi from "u2f-api";
import BluebirdPromise = require("bluebird");
import { SignMessage } from "../../../../shared/SignMessage";
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

function startU2fAuthentication($: JQueryStatic, notifier: INotifier)
  : BluebirdPromise<string> {

  return GetPromised($, Endpoints.SECOND_FACTOR_U2F_SIGN_REQUEST_GET, {},
    undefined, "json")
    .then(function (signRequest: U2f.Request) {
      notifier.info(UserMessages.PLEASE_TOUCH_TOKEN);
      return U2fApi.sign(signRequest, 60);
    })
    .then(function (signResponse: U2fApi.SignResponse) {
      return finishU2fAuthentication(signResponse, $);
    });
}

export function validate($: JQueryStatic, notifier: INotifier) {
  return startU2fAuthentication($, notifier)
    .catch(function (err: Error) {
      notifier.error(UserMessages.U2F_TRANSACTION_FINISH_FAILED);
      return BluebirdPromise.reject(err);
    });
}

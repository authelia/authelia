
import U2fApi = require("u2f-api");
import U2f = require("u2f");
import BluebirdPromise = require("bluebird");
import { SignMessage } from "../../../../shared/SignMessage";
import Endpoints = require("../../../../shared/api");
import UserMessages = require("../../../../shared/UserMessages");
import { INotifier } from "../INotifier";
import { RedirectionMessage } from "../../../../shared/RedirectionMessage";
import { ErrorMessage } from "../../../../shared/ErrorMessage";

function finishU2fAuthentication(responseData: U2fApi.SignResponse, $: JQueryStatic): BluebirdPromise<string> {
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

function startU2fAuthentication($: JQueryStatic, notifier: INotifier, u2fApi: typeof U2fApi): BluebirdPromise<string> {
  return new BluebirdPromise<string>(function (resolve, reject) {
    $.get(Endpoints.SECOND_FACTOR_U2F_SIGN_REQUEST_GET, {}, undefined, "json")
      .done(function (signResponse: SignMessage) {
        notifier.info(UserMessages.PLEASE_TOUCH_TOKEN);

        const signRequest: U2fApi.SignRequest = {
          appId: signResponse.request.appId,
          challenge: signResponse.request.challenge,
          keyHandle: signResponse.keyHandle, // linked to the client session cookie
          version: "U2F_V2"
        };

        u2fApi.sign([signRequest], 60)
          .then(function (signResponse: U2fApi.SignResponse) {
            finishU2fAuthentication(signResponse, $)
              .then(function (redirect: string) {
                resolve(redirect);
              }, function (err) {
                notifier.error(UserMessages.U2F_TRANSACTION_FINISH_FAILED);
                reject(err);
              });
          })
          .catch(function (err: Error) {
            reject(err);
          });
      })
      .fail(function (xhr: JQueryXHR, textStatus: string) {
        reject(new Error(textStatus));
      });
  });
}


export function validate($: JQueryStatic, notifier: INotifier, u2fApi: typeof U2fApi): BluebirdPromise<string> {
  return startU2fAuthentication($, notifier, u2fApi);
}

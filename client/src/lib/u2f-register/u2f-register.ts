
import BluebirdPromise = require("bluebird");
import U2f = require("u2f");
import U2fApi = require("u2f-api-polyfill");
import jslogger = require("js-logger");
import { Notifier } from "../Notifier";
import GetPromised from "../GetPromised";
import Endpoints = require("../../../../shared/api");
import UserMessages = require("../../../../shared/UserMessages");
import { RedirectionMessage } from "../../../../shared/RedirectionMessage";
import { ErrorMessage } from "../../../../shared/ErrorMessage";

export default function (window: Window, $: JQueryStatic) {
  const notifier = new Notifier(".notification", $);

  function checkRegistration(regResponse: U2fApi.RegisterResponse): BluebirdPromise<string> {
    return new BluebirdPromise<string>(function (resolve, reject) {
      $.post(Endpoints.SECOND_FACTOR_U2F_REGISTER_POST, regResponse, undefined, "json")
        .done((body: RedirectionMessage | ErrorMessage) => {
          if (body && "error" in body) {
            reject(new Error((body as ErrorMessage).error));
            return;
          }
          resolve((body as RedirectionMessage).redirect);
        })
        .fail((xhr, status) => {
          reject(new Error("Failed to register device."));
        });
    });
  }

  function register(appId: string, registerRequest: U2fApi.RegisterRequest,
    timeout: number): BluebirdPromise<U2fApi.RegisterResponse> {
    return new BluebirdPromise((resolve, reject) => {
      (window as any).u2f.register(appId, [registerRequest], [],
        (res: U2fApi.RegisterResponse | U2fApi.U2FError) => {
          if ((<U2fApi.U2FError>res).errorCode != 0) {
            reject(new Error((<U2fApi.U2FError>res).errorMessage));
            return;
          }
          resolve(<U2fApi.RegisterResponse>res);
        }, timeout);
    });
  }

  function requestRegistration(): BluebirdPromise<string> {
    return GetPromised($, Endpoints.SECOND_FACTOR_U2F_REGISTER_REQUEST_GET, {},
      undefined, "json")
      .then((registrationRequest: U2f.Request) => {
        return register(registrationRequest.appId, registrationRequest, 60);
      })
      .then((res) => checkRegistration(res));
  }

  function onRegisterFailure(err: Error) {
    notifier.error(UserMessages.REGISTRATION_U2F_FAILED);
  }

  $(document).ready(function () {
    requestRegistration()
      .then((redirectionUrl: string) => {
        document.location.href = redirectionUrl;
      })
      .catch((err) => {
        onRegisterFailure(err);
      });
  });
}

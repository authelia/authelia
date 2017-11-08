
import BluebirdPromise = require("bluebird");
import U2f = require("u2f");
import U2fApi = require("u2f-api");
import jslogger = require("js-logger");
import { Notifier } from "../Notifier";
import GetPromised from "../GetPromised";
import Endpoints = require("../../../../shared/api");
import UserMessages = require("../../../../shared/UserMessages");
import { RedirectionMessage } from "../../../../shared/RedirectionMessage";
import { ErrorMessage } from "../../../../shared/ErrorMessage";

export default function (window: Window, $: JQueryStatic, u2fApi: U2fApi.U2fApi) {
  const notifier = new Notifier(".notification", $);

  function checkRegistration(regResponse: U2fApi.RegisterResponse): BluebirdPromise<string> {
    const registrationData: U2f.RegistrationData = regResponse;

    return new BluebirdPromise<string>(function (resolve, reject) {
      $.post(Endpoints.SECOND_FACTOR_U2F_REGISTER_POST, registrationData, undefined, "json")
        .done(function (body: RedirectionMessage | ErrorMessage) {
          if (body && "error" in body) {
            reject(new Error((body as ErrorMessage).error));
            return;
          }
          resolve((body as RedirectionMessage).redirect);
        })
        .fail(function (xhr, status) {
          reject();
        });
    });
  }

  function u2fApiRegister(u2fApi: U2fApi.U2fApi, appId: string,
    registerRequest: U2fApi.RegisterRequest, timeout: number) {

    return new BluebirdPromise(function (resolve, reject) {
      u2fApi.register(appId, [registerRequest], [],
        function (res: U2fApi.RegisterResponse | U2fApi.Error) {
          if ("errorCode" in res) {
            reject(new Error((res as U2fApi.Error).errorMessage));
            return;
          }
          resolve(res);
        }, timeout);
    });
  }

  function requestRegistration(): BluebirdPromise<string> {
    return GetPromised($, Endpoints.SECOND_FACTOR_U2F_REGISTER_REQUEST_GET, {},
      undefined, "json")
      .then(function (registrationRequest: U2f.Request) {
        const registerRequest: U2fApi.RegisterRequest = registrationRequest;
        const appId = registrationRequest.appId;
        return u2fApiRegister(u2fApi, appId, registerRequest, 60);
      })
      .then(function (res: U2fApi.RegisterResponse) {
        return checkRegistration(res);
      });
  }

  function onRegisterFailure(err: Error) {
    notifier.error(UserMessages.REGISTRATION_U2F_FAILED);
  }

  $(document).ready(function () {
    requestRegistration()
      .then(function (redirectionUrl: string) {
        document.location.href = redirectionUrl;
      })
      .error(function (err) {
        onRegisterFailure(err);
      });
  });
}


import BluebirdPromise = require("bluebird");
import U2f = require("u2f");
import u2fApi = require("u2f-api");
import jslogger = require("js-logger");
import { Notifier } from "../Notifier";
import Endpoints = require("../../../../shared/api");
import UserMessages = require("../../../../shared/UserMessages");

export default function (window: Window, $: JQueryStatic) {
  const notifier = new Notifier(".notification", $);

  function checkRegistration(regResponse: u2fApi.RegisterResponse): BluebirdPromise<string> {
    const registrationData: U2f.RegistrationData = regResponse;

    jslogger.debug("registrationResponse = %s", JSON.stringify(registrationData));

    return new BluebirdPromise<string>(function (resolve, reject) {
      $.post(Endpoints.SECOND_FACTOR_U2F_REGISTER_POST, registrationData, undefined, "json")
        .done(function (body: any) {
          if (body && body.error) {
            reject(new Error(body.error));
            return;
          }
          resolve(body.redirection_url);
        })
        .fail(function (xhr, status) {
          reject();
        });
    });
  }

  function requestRegistration(): BluebirdPromise<string> {
    return new BluebirdPromise<string>(function (resolve, reject) {
      $.get(Endpoints.SECOND_FACTOR_U2F_REGISTER_REQUEST_GET, {}, undefined, "json")
        .done(function (registrationRequest: U2f.Request) {
          const registerRequest: u2fApi.RegisterRequest = registrationRequest;
          u2fApi.register([registerRequest], [], 120)
            .then(function (res: u2fApi.RegisterResponse) {
              return checkRegistration(res);
            })
            .then(function (redirectionUrl: string) {
              resolve(redirectionUrl);
            })
            .catch(function (err: Error) {
              reject(err);
            });
        });
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


import BluebirdPromise = require("bluebird");
import U2f = require("u2f");
import * as U2fApi from "u2f-api";
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

  function requestRegistration(): BluebirdPromise<string> {
    return GetPromised($, Endpoints.SECOND_FACTOR_U2F_REGISTER_REQUEST_GET, {},
      undefined, "json")
      .then((registrationRequest: U2f.Request) => {
        return U2fApi.register(registrationRequest, [], 60);
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

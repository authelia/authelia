
import BluebirdPromise = require("bluebird");

import Endpoints = require("../../../../shared/api");
import UserMessages = require("../../../../shared/UserMessages");
import Constants = require("./constants");
import jslogger = require("js-logger");
import { Notifier } from "../Notifier";

export default function (window: Window, $: JQueryStatic) {
  const notifier = new Notifier(".notification", $);

  function requestPasswordReset(username: string) {
    return new BluebirdPromise(function (resolve, reject) {
      $.get(Endpoints.RESET_PASSWORD_IDENTITY_START_GET, {
        userid: username,
      })
        .done(function (body: any) {
          if (body && body.error) {
            reject(new Error(body.error));
            return;
          }
          resolve();
        })
        .fail(function (xhr: JQueryXHR, textStatus: string) {
          reject(new Error(textStatus));
        });
    });
  }

  function onFormSubmitted() {
    const username = $("#username").val() as string;

    if (!username) {
      notifier.warning(UserMessages.MISSING_USERNAME);
      return;
    }

    requestPasswordReset(username)
      .then(function () {
        notifier.success(UserMessages.MAIL_SENT);
        setTimeout(function () {
          window.location.replace(Endpoints.FIRST_FACTOR_GET);
        }, 1000);
      })
      .error(function () {
        notifier.error(UserMessages.MAIL_NOT_SENT);
      });
    return false;
  }

  $(document).ready(function () {
    $(Constants.FORM_SELECTOR).on("submit", onFormSubmitted);
  });
}



import BluebirdPromise = require("bluebird");

import Endpoints = require("../../../../shared/api");
import Constants = require("./constants");
import jslogger = require("js-logger");
import { Notifier } from "../Notifier";

export default function(window: Window, $: JQueryStatic) {
  const notifier = new Notifier(".notification", $);

  function requestPasswordReset(username: string) {
    return new BluebirdPromise(function (resolve, reject) {
      $.get(Endpoints.RESET_PASSWORD_IDENTITY_START_GET, {
        userid: username,
      })
        .done(function () {
          resolve();
        })
        .fail(function (xhr: JQueryXHR, textStatus: string) {
          reject(new Error(textStatus));
        });
    });
  }

  function onFormSubmitted() {
    const username = $("#username").val();

    if (!username) {
      notifier.warning("You must provide your username to reset your password.");
      return;
    }

    requestPasswordReset(username)
      .then(function () {
        notifier.success("An email has been sent to you. Follow the link to change your password.");
        setTimeout(function () {
          window.location.replace(Endpoints.FIRST_FACTOR_GET);
        }, 1000);
      })
      .error(function () {
        notifier.warning("Are you sure this is your username?");
      });
      return false;
  }

  $(document).ready(function () {
    jslogger.debug("Reset password request form setup");
    $(Constants.FORM_SELECTOR).on("submit", onFormSubmitted);
  });
}


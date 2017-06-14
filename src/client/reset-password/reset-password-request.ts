
import BluebirdPromise = require("bluebird");

import Endpoints = require("../../server/endpoints");
import Constants = require("./constants");
import jslogger = require("js-logger");

export default function(window: Window, $: JQueryStatic) {
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
      $.notify("You must provide your username to reset your password.", "warn");
      return;
    }

    requestPasswordReset(username)
      .then(function () {
        $.notify("An email has been sent. Click on the link to change your password", "success");
        setTimeout(function () {
          window.location.replace(Endpoints.FIRST_FACTOR_GET);
        }, 1000);
      })
      .error(function () {
        $.notify("Are you sure this is your username?", "warn");
      });
      return false;
  }

  $(document).ready(function () {
    jslogger.debug("Reset password request form setup");
    $(Constants.FORM_SELECTOR).on("submit", onFormSubmitted);
  });
}


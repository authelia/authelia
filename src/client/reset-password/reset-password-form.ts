import BluebirdPromise = require("bluebird");

import Endpoints = require("../../server/endpoints");
import Constants = require("./constants");

export default function (window: Window, $: JQueryStatic) {
  function modifyPassword(newPassword: string) {
    return new BluebirdPromise(function (resolve, reject) {
      $.post(Endpoints.RESET_PASSWORD_FORM_POST, {
        password: newPassword,
      })
        .done(function (data) {
          resolve(data);
        })
        .fail(function (xhr, status) {
          reject(status);
        });
    });
  }

  function onFormSubmitted() {
    const password1 = $("#password1").val();
    const password2 = $("#password2").val();

    if (!password1 || !password2) {
      $.notify("You must enter your new password twice.", "warn");
      return false;
    }

    if (password1 != password2) {
      $.notify("The passwords are different", "warn");
      return false;
    }

    modifyPassword(password1)
      .then(function () {
        $.notify("Your password has been changed. Please login again", "success");
        window.location.href = Endpoints.FIRST_FACTOR_GET;
      })
      .error(function () {
        $.notify("An error occurred during password change.", "warn");
      });
    return false;
  }

  $(document).ready(function () {
    $(Constants.FORM_SELECTOR).on("submit", onFormSubmitted);
  });
}

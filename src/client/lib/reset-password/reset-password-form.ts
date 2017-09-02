import BluebirdPromise = require("bluebird");

import Endpoints = require("../../../server/endpoints");
import Constants = require("./constants");
import { Notifier } from "../Notifier";

export default function (window: Window, $: JQueryStatic) {
  const notifier = new Notifier(".notification", $);

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
      notifier.warning("You must enter your new password twice.");
      return false;
    }

    if (password1 != password2) {
      notifier.warning("The passwords are different.");
      return false;
    }

    modifyPassword(password1)
      .then(function () {
        notifier.success("Your password has been changed. Please log in again.");
        window.location.href = Endpoints.FIRST_FACTOR_GET;
      })
      .error(function () {
        notifier.warning("An error occurred during password reset. Your password has not been changed.");
      });
    return false;
  }

  $(document).ready(function () {
    $(Constants.FORM_SELECTOR).on("submit", onFormSubmitted);
  });
}

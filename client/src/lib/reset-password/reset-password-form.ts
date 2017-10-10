import BluebirdPromise = require("bluebird");

import Endpoints = require("../../../../shared/api");
import UserMessages = require("../../../../shared/UserMessages");

import Constants = require("./constants");
import { Notifier } from "../Notifier";

export default function (window: Window, $: JQueryStatic) {
  const notifier = new Notifier(".notification", $);

  function modifyPassword(newPassword: string) {
    return new BluebirdPromise(function (resolve, reject) {
      $.post(Endpoints.RESET_PASSWORD_FORM_POST, {
        password: newPassword,
      })
        .done(function (body: any) {
          if (body && body.error) {
            reject(new Error(body.error));
            return;
          }
          resolve(body);
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
      notifier.warning(UserMessages.MISSING_PASSWORD);
      return false;
    }

    if (password1 != password2) {
      notifier.warning(UserMessages.DIFFERENT_PASSWORDS);
      return false;
    }

    modifyPassword(password1)
      .then(function () {
        window.location.href = Endpoints.FIRST_FACTOR_GET;
      })
      .error(function () {
        notifier.error(UserMessages.RESET_PASSWORD_FAILED);
      });
    return false;
  }

  $(document).ready(function () {
    $(Constants.FORM_SELECTOR).on("submit", onFormSubmitted);
  });
}

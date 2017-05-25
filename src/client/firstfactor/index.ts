import FirstFactorValidator = require("./FirstFactorValidator");
import JSLogger = require("js-logger");
import UISelectors = require("./UISelectors");

import Endpoints = require("../../server/endpoints");

export default function (window: Window, $: JQueryStatic, firstFactorValidator: typeof FirstFactorValidator, jslogger: typeof JSLogger) {
  function onFormSubmitted() {
    const username: string = $(UISelectors.USERNAME_FIELD_ID).val();
    const password: string = $(UISelectors.PASSWORD_FIELD_ID).val();
    jslogger.debug("Form submitted");
    firstFactorValidator.validate(username, password, $)
      .then(onFirstFactorSuccess, onFirstFactorFailure);
    return false;
  }

  function onFirstFactorSuccess() {
    jslogger.debug("First factor validated.");
    $(UISelectors.USERNAME_FIELD_ID).val("");
    $(UISelectors.PASSWORD_FIELD_ID).val("");

    // Redirect to second factor
    window.location.href = Endpoints.SECOND_FACTOR_GET;
  }

  function onFirstFactorFailure(err: Error) {
    jslogger.debug("First factor failed.");

    $(UISelectors.PASSWORD_FIELD_ID).val("");
    $.notify("Error during authentication: " + err.message, "error");
  }


  $(window.document).ready(function () {
    jslogger.info("Enter first factor");
    $("form").on("submit", onFormSubmitted);
  });
}


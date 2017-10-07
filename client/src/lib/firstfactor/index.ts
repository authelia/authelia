import FirstFactorValidator = require("./FirstFactorValidator");
import JSLogger = require("js-logger");
import UISelectors = require("./UISelectors");
import { Notifier } from "../Notifier";
import { QueryParametersRetriever } from "../QueryParametersRetriever";
import Constants = require("../../../../shared/constants");
import Endpoints = require("../../../../shared/api");

export default function (window: Window, $: JQueryStatic,
  firstFactorValidator: typeof FirstFactorValidator, jslogger: typeof JSLogger) {

  const notifier = new Notifier(".notification", $);

  function onFormSubmitted() {
    const username: string = $(UISelectors.USERNAME_FIELD_ID).val();
    const password: string = $(UISelectors.PASSWORD_FIELD_ID).val();
    $(UISelectors.PASSWORD_FIELD_ID).val("");

    const redirectUrl = QueryParametersRetriever.get(Constants.REDIRECT_QUERY_PARAM);
    const onlyBasicAuth = QueryParametersRetriever.get(Constants.ONLY_BASIC_AUTH_QUERY_PARAM) ? true : false;
    firstFactorValidator.validate(username, password, redirectUrl, onlyBasicAuth, $)
      .then(onFirstFactorSuccess, onFirstFactorFailure);
    return false;
  }

  function onFirstFactorSuccess(redirectUrl: string) {
    jslogger.debug("First factor validated.");
      window.location.href = redirectUrl;
  }

  function onFirstFactorFailure(err: Error) {
    jslogger.debug("First factor failed.");
    notifier.error("Authentication failed. Please double check your credentials.");
  }


  $(window.document).ready(function () {
    jslogger.info("Enter first factor");
    $("form").on("submit", onFormSubmitted);
  });
}


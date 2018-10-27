import FirstFactorValidator = require("./FirstFactorValidator");
import JSLogger = require("js-logger");
import UISelectors = require("./UISelectors");
import { Notifier } from "../Notifier";
import { QueryParametersRetriever } from "../QueryParametersRetriever";
import Constants = require("../../../../shared/constants");
import Endpoints = require("../../../../shared/api");
import UserMessages = require("../../../../shared/UserMessages");
import { SafeRedirect } from "../SafeRedirect";

export default function (window: Window, $: JQueryStatic,
  firstFactorValidator: typeof FirstFactorValidator, jslogger: typeof JSLogger) {

  const notifier = new Notifier(".notification", $);

  function onFormSubmitted() {
    const username: string = $(UISelectors.USERNAME_FIELD_ID).val() as string;
    const password: string = $(UISelectors.PASSWORD_FIELD_ID).val() as string;
    const keepMeLoggedIn: boolean = $(UISelectors.KEEP_ME_LOGGED_IN_ID).is(":checked");

    $("form").css("opacity", 0.5);
    $("input,button").attr("disabled", "true");
    $(UISelectors.SIGN_IN_BUTTON_ID).text("Please wait...");

    const redirectUrl = QueryParametersRetriever.get(Constants.REDIRECT_QUERY_PARAM);
    firstFactorValidator.validate(username, password, keepMeLoggedIn, redirectUrl, $)
      .then(onFirstFactorSuccess, onFirstFactorFailure);
    return false;
  }

  function onFirstFactorSuccess(redirectUrl: string) {
    SafeRedirect(redirectUrl, () => {
      notifier.error("Cannot redirect to an external domain.");
    });
  }

  function onFirstFactorFailure(err: Error) {
    $("input,button").removeAttr("disabled");
    $("form").css("opacity", 1);
    notifier.error(UserMessages.AUTHENTICATION_FAILED);
    $(UISelectors.PASSWORD_FIELD_ID).select();
    $(UISelectors.SIGN_IN_BUTTON_ID).text("Sign in");
  }

  $(window.document).ready(function () {
    $("form").on("submit", onFormSubmitted);
  });
}


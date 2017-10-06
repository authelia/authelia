
import U2fApi = require("u2f-api");
import jslogger = require("js-logger");

import TOTPValidator = require("./TOTPValidator");
import U2FValidator = require("./U2FValidator");
import Constants = require("./constants");
import { Notifier } from "../Notifier";
import { QueryParametersRetriever } from "../QueryParametersRetriever";
import Endpoints = require("../../../../shared/api");
import ServerConstants = require("../../../../shared/constants");


export default function (window: Window, $: JQueryStatic, u2fApi: typeof U2fApi) {
  const notifierTotp = new Notifier(".notification-totp", $);
  const notifierU2f = new Notifier(".notification-u2f", $);

  function onAuthenticationSuccess(data: any, notifier: Notifier) {
    const redirectUrl = QueryParametersRetriever.get(ServerConstants.REDIRECT_QUERY_PARAM);
    if (redirectUrl)
      window.location.href = redirectUrl;
    else
      notifier.success("Authentication succeeded. You can now access your services.");
  }

  function onSecondFactorTotpSuccess(data: any) {
    onAuthenticationSuccess(data, notifierTotp);
  }

  function onSecondFactorTotpFailure(err: Error) {
    notifierTotp.error("Problem with TOTP validation.");
  }

  function onU2fAuthenticationSuccess(data: any) {
    onAuthenticationSuccess(data, notifierU2f);
  }

  function onU2fAuthenticationFailure() {
    notifierU2f.error("Problem with U2F validation. Did you register before authenticating?");
  }

  function onTOTPFormSubmitted(): boolean {
    const token = $(Constants.TOTP_TOKEN_SELECTOR).val();
    jslogger.debug("TOTP token is %s", token);

    TOTPValidator.validate(token, $)
      .then(onSecondFactorTotpSuccess)
      .catch(onSecondFactorTotpFailure);
    return false;
  }

  function onU2FFormSubmitted(): boolean {
    jslogger.debug("Start U2F authentication");
    U2FValidator.validate($, notifierU2f, U2fApi)
      .then(onU2fAuthenticationSuccess, onU2fAuthenticationFailure);
    return false;
  }

  $(window.document).ready(function () {
    $(Constants.TOTP_FORM_SELECTOR).on("submit", onTOTPFormSubmitted);
    $(Constants.U2F_FORM_SELECTOR).on("submit", onU2FFormSubmitted);
  });
}
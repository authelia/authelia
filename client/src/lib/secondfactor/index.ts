import jslogger = require("js-logger");
import U2fApi = require("../../../types/u2f-api");
import TOTPValidator = require("./TOTPValidator");
import U2FValidator = require("./U2FValidator");
import ClientConstants = require("./constants");
import { Notifier } from "../Notifier";
import { QueryParametersRetriever } from "../QueryParametersRetriever";
import Endpoints = require("../../../../shared/api");
import ServerConstants = require("../../../../shared/constants");
import UserMessages = require("../../../../shared/UserMessages");
import SharedConstants = require("../../../../shared/constants");

export default function (window: Window, $: JQueryStatic, u2fApi: U2fApi.U2fApi) {
  const notifierTotp = new Notifier(".notification-totp", $);
  const notifierU2f = new Notifier(".notification-u2f", $);

  function onAuthenticationSuccess(serverRedirectUrl: string, notifier: Notifier) {
    if (QueryParametersRetriever.get(SharedConstants.REDIRECT_QUERY_PARAM))
      window.location.href = QueryParametersRetriever.get(SharedConstants.REDIRECT_QUERY_PARAM);
    else if (serverRedirectUrl)
      window.location.href = serverRedirectUrl;
    else
      notifier.success(UserMessages.AUTHENTICATION_SUCCEEDED);
  }

  function onSecondFactorTotpSuccess(redirectUrl: string) {
    onAuthenticationSuccess(redirectUrl, notifierTotp);
  }

  function onSecondFactorTotpFailure(err: Error) {
    notifierTotp.error(UserMessages.AUTHENTICATION_TOTP_FAILED);
  }

  function onU2fAuthenticationSuccess(redirectUrl: string) {
    onAuthenticationSuccess(redirectUrl, notifierU2f);
  }

  function onU2fAuthenticationFailure() {
    notifierU2f.error(UserMessages.AUTHENTICATION_U2F_FAILED);
  }

  function onTOTPFormSubmitted(): boolean {
    const token = $(ClientConstants.TOTP_TOKEN_SELECTOR).val() as string;
    TOTPValidator.validate(token, $)
      .then(onSecondFactorTotpSuccess)
      .catch(onSecondFactorTotpFailure);
    return false;
  }

  $(window.document).ready(function () {
    $(ClientConstants.TOTP_FORM_SELECTOR).on("submit", onTOTPFormSubmitted);
    U2FValidator.validate($, notifierU2f, u2fApi)
      .then(onU2fAuthenticationSuccess, onU2fAuthenticationFailure);
  });
}
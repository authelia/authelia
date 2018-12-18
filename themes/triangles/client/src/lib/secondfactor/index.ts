import TOTPValidator = require("./TOTPValidator");
import U2FValidator = require("./U2FValidator");
import ClientConstants = require("./constants");
import { Notifier } from "../Notifier";
import { QueryParametersRetriever } from "../QueryParametersRetriever";
import UserMessages = require("../../../../shared/UserMessages");
import SharedConstants = require("../../../../shared/constants");
import { SafeRedirect } from "../SafeRedirect";

export default function (window: Window, $: JQueryStatic) {
  const notifier = new Notifier(".notification", $);

  function onAuthenticationSuccess(serverRedirectUrl: string) {
    const queryRedirectUrl = QueryParametersRetriever.get(SharedConstants.REDIRECT_QUERY_PARAM);
    if (queryRedirectUrl) {
      SafeRedirect(queryRedirectUrl, () => {
        notifier.error(UserMessages.CANNOT_REDIRECT_TO_EXTERNAL_DOMAIN);
      });
    } else if (serverRedirectUrl) {
      SafeRedirect(serverRedirectUrl, () => {
        notifier.error(UserMessages.CANNOT_REDIRECT_TO_EXTERNAL_DOMAIN);
      });
    } else {
      notifier.success(UserMessages.AUTHENTICATION_SUCCEEDED);
    }
  }

  function onSecondFactorTotpSuccess(redirectUrl: string) {
    onAuthenticationSuccess(redirectUrl);
  }

  function onSecondFactorTotpFailure(err: Error) {
    notifier.error(UserMessages.AUTHENTICATION_TOTP_FAILED);
  }

  function onU2fAuthenticationSuccess(redirectUrl: string) {
    onAuthenticationSuccess(redirectUrl);
  }

  function onU2fAuthenticationFailure() {
    // TODO(clems4ever): we should not display this error message until a device
    //                   is registered.
    // notifier.error(UserMessages.AUTHENTICATION_U2F_FAILED);
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
    U2FValidator.validate($)
      .then(onU2fAuthenticationSuccess, onU2fAuthenticationFailure);
  });
}

import U2fApi = require("u2f-api");
import jslogger = require("js-logger");

import TOTPValidator = require("./TOTPValidator");
import U2FValidator = require("./U2FValidator");
import Endpoints = require("../../../server/endpoints");
import Constants = require("./constants");
import { Notifier } from "../Notifier";
import { QueryParametersRetriever } from "../QueryParametersRetriever";


export default function (window: Window, $: JQueryStatic, u2fApi: typeof U2fApi) {
    const notifierTotp = new Notifier(".notification-totp", $);
    const notifierU2f = new Notifier(".notification-u2f", $);

    function onAuthenticationSuccess(data: any) {
        const redirectUrl = QueryParametersRetriever.get("redirect");
        if (redirectUrl)
            window.location.href = redirectUrl;
        else
            window.location.href = Endpoints.FIRST_FACTOR_GET;
    }

    function onSecondFactorTotpSuccess(data: any) {
        onAuthenticationSuccess(data);
    }

    function onSecondFactorTotpFailure(err: Error) {
        notifierTotp.error("Problem with TOTP validation.");
    }

    function onU2fAuthenticationSuccess(data: any) {
        onAuthenticationSuccess(data);
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
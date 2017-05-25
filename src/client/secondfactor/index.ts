
import U2fApi = require("u2f-api");
import jslogger = require("js-logger");

import TOTPValidator = require("./TOTPValidator");
import U2FValidator = require("./U2FValidator");

import Endpoints = require("../../server/endpoints");

import Constants = require("./constants");


export default function (window: Window, $: JQueryStatic, u2fApi: typeof U2fApi) {
    function onAuthenticationSuccess(data: any) {
        window.location.href = data.redirection_url;
    }


    function onSecondFactorTotpSuccess(data: any) {
        onAuthenticationSuccess(data);
    }

    function onSecondFactorTotpFailure(err: Error) {
        $.notify("Error while validating TOTP token. Cause: " + err.message, "error");
    }

    function onU2fAuthenticationSuccess(data: any) {
        onAuthenticationSuccess(data);
    }

    function onU2fAuthenticationFailure() {
        $.notify("Problem with U2F authentication. Did you register before authenticating?", "warn");
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
        U2FValidator.validate($, U2fApi)
            .then(onU2fAuthenticationSuccess, onU2fAuthenticationFailure);
        return false;
    }

    $(window.document).ready(function () {
        $(Constants.TOTP_FORM_SELECTOR).on("submit", onTOTPFormSubmitted);
        $(Constants.U2F_FORM_SELECTOR).on("submit", onU2FFormSubmitted);
    });
}
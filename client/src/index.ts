
import FirstFactorValidator = require("./lib/firstfactor/FirstFactorValidator");

import FirstFactor from "./lib/firstfactor/index";
import SecondFactor from "./lib/secondfactor/index";
import TOTPRegister from "./lib/totp-register/totp-register";
import U2fRegister from "./lib/u2f-register/u2f-register";
import ResetPasswordRequest from "./lib/reset-password/reset-password-request";
import ResetPasswordForm from "./lib/reset-password/reset-password-form";
import jslogger = require("js-logger");
import jQuery = require("jquery");
import U2fApi = require("u2f-api");

jslogger.useDefaults();
jslogger.setLevel(jslogger.INFO);

export = {
    firstfactor: function () {
        FirstFactor(window, jQuery, FirstFactorValidator, jslogger);
    },
    secondfactor: function () {
        SecondFactor(window, jQuery, U2fApi);
    },
    register_totp: function() {
        TOTPRegister(window, jQuery);
    },
    register_u2f: function () {
        U2fRegister(window, jQuery);
    },
    reset_password_request: function () {
        ResetPasswordRequest(window, jQuery);
    },
    reset_password_form: function () {
        ResetPasswordForm(window, jQuery);
    }
};
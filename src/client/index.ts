
import FirstFactorValidator = require("./firstfactor/FirstFactorValidator");

import FirstFactor from "./firstfactor/index";
import SecondFactor from "./secondfactor/index";
import TOTPRegister from "./totp-register/totp-register";
import U2fRegister from "./u2f-register/u2f-register";
import ResetPasswordRequest from "./reset-password/reset-password-request";
import ResetPasswordForm from "./reset-password/reset-password-form";
import jslogger = require("js-logger");
import jQuery = require("jquery");
import u2fApi = require("u2f-api");

jslogger.useDefaults();
jslogger.setLevel(jslogger.INFO);

require("notifyjs-browser")(jQuery);

export = {
    firstfactor: function () {
        FirstFactor(window, jQuery, FirstFactorValidator, jslogger);
    },
    secondfactor: function () {
        SecondFactor(window, jQuery, u2fApi);
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

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
import Endpoints = require("../../shared/api");

jslogger.useDefaults();
jslogger.setLevel(jslogger.INFO);

(function () {
  if (window.location.pathname == Endpoints.FIRST_FACTOR_GET)
    FirstFactor(window, jQuery, FirstFactorValidator, jslogger);
  else if (window.location.pathname == Endpoints.SECOND_FACTOR_GET)
    SecondFactor(window, jQuery, U2fApi);
  else if (window.location.pathname == Endpoints.SECOND_FACTOR_TOTP_IDENTITY_FINISH_GET)
    TOTPRegister(window, jQuery);
  else if (window.location.pathname == Endpoints.SECOND_FACTOR_U2F_IDENTITY_FINISH_GET)
    U2fRegister(window, jQuery);
  else if (window.location.pathname == Endpoints.RESET_PASSWORD_IDENTITY_FINISH_GET)
    ResetPasswordForm(window, jQuery);
  else if (window.location.pathname == Endpoints.RESET_PASSWORD_REQUEST_GET)
    ResetPasswordRequest(window, jQuery);
})();


import FirstFactorValidator = require("./lib/firstfactor/FirstFactorValidator");

import FirstFactor from "./lib/firstfactor/index";
import SecondFactor from "./lib/secondfactor/index";
import TOTPRegister from "./lib/totp-register/totp-register";
import U2fRegister from "./lib/u2f-register/u2f-register";
import ResetPasswordRequest from "./lib/reset-password/reset-password-request";
import ResetPasswordForm from "./lib/reset-password/reset-password-form";
import jslogger = require("js-logger");
import jQuery = require("jquery");
import Endpoints = require("../../shared/api");

jslogger.useDefaults();
jslogger.setLevel(jslogger.INFO);

(function () {
  (<any>window).jQuery = jQuery;
  require("bootstrap");

  jQuery('[data-toggle="tooltip"]').tooltip();
  if (window.location.pathname == Endpoints.FIRST_FACTOR_GET)
    FirstFactor(window, jQuery, FirstFactorValidator, jslogger);
  else if (window.location.pathname == Endpoints.SECOND_FACTOR_GET)
    SecondFactor(window, jQuery, (global as any).u2f);
  else if (window.location.pathname == Endpoints.SECOND_FACTOR_TOTP_IDENTITY_FINISH_GET)
    TOTPRegister(window, jQuery);
  else if (window.location.pathname == Endpoints.SECOND_FACTOR_U2F_IDENTITY_FINISH_GET)
    U2fRegister(window, jQuery, (global as any).u2f);
  else if (window.location.pathname == Endpoints.RESET_PASSWORD_IDENTITY_FINISH_GET)
    ResetPasswordForm(window, jQuery);
  else if (window.location.pathname == Endpoints.RESET_PASSWORD_REQUEST_GET)
    ResetPasswordRequest(window, jQuery);
})();

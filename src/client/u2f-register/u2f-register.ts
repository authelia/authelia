
import BluebirdPromise = require("bluebird");
import U2f = require("u2f");
import u2fApi = require("u2f-api");

import Endpoints = require("../../server/endpoints");
import jslogger = require("js-logger");

export default function(window: Window, $: JQueryStatic) {

  function checkRegistration(regResponse: u2fApi.RegisterResponse, fn: (err: Error) => void) {
    const registrationData: U2f.RegistrationData = regResponse;

    jslogger.debug("registrationResponse = %s", JSON.stringify(registrationData));

    $.post(Endpoints.SECOND_FACTOR_U2F_REGISTER_POST, registrationData, undefined, "json")
      .done(function (data) {
        document.location.href = data.redirection_url;
      })
      .fail(function (xhr, status) {
        $.notify("Error when finish U2F transaction" + status);
      });
  }

  function requestRegistration(fn: (err: Error) => void) {
    $.get(Endpoints.SECOND_FACTOR_U2F_REGISTER_REQUEST_GET, {}, undefined, "json")
      .done(function (registrationRequest: U2f.Request) {
        jslogger.debug("registrationRequest = %s", JSON.stringify(registrationRequest));

        const registerRequest: u2fApi.RegisterRequest = registrationRequest;
        u2fApi.register([registerRequest], [], 120)
          .then(function (res: u2fApi.RegisterResponse) {
            checkRegistration(res, fn);
          })
          .catch(function (err: Error) {
            fn(err);
          });
      });
  }

  function onRegisterFailure(err: Error) {
    $.notify("Problem authenticating with U2F.", "error");
  }

  $(document).ready(function () {
    requestRegistration(function (err: Error) {
      if (err) {
        onRegisterFailure(err);
        return;
      }
    });
  });
}

import SecondFactorGet from "../../../src/lib/routes/secondfactor/get";
import { ServerVariablesMockBuilder, ServerVariablesMock }
  from "../../mocks/ServerVariablesMockBuilder";
import { ServerVariables } from "../../../src/lib/ServerVariables";
import Sinon = require("sinon");
import ExpressMock = require("../../mocks/express");
import Assert = require("assert");
import Endpoints = require("../../../../shared/api");
import BluebirdPromise = require("bluebird");

describe("test second factor GET endpoint handler", function () {
  let mocks: ServerVariablesMock;
  let vars: ServerVariables;
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;

  beforeEach(function () {
    const s = ServerVariablesMockBuilder.build();
    mocks = s.mocks;
    vars = s.variables;

    req = ExpressMock.RequestMock();
    res = ExpressMock.ResponseMock();

    req.session = {
      auth: {
        userid: "user",
        first_factor: true,
        second_factor: false
      }
    };
  });

  describe("test redirection", function () {
    it("should redirect to already logged in page if server is in single_factor mode", function () {
      vars.config.authentication_methods.default_method = "single_factor";
      return SecondFactorGet(vars)(req as any, res as any)
        .then(function () {
          Assert(res.redirect.calledWith(Endpoints.LOGGED_IN));
          return BluebirdPromise.resolve();
        });
    });

    it("should redirect to already logged in page if user already authenticated", function () {
      vars.config.authentication_methods.default_method = "two_factor";
      req.session.auth.second_factor = true;
      return SecondFactorGet(vars)(req as any, res as any)
        .then(function () {
          Assert(res.redirect.calledWith(Endpoints.LOGGED_IN));
          return BluebirdPromise.resolve();
        });
    });

    it("should render second factor page", function () {
      vars.config.authentication_methods.default_method = "two_factor";
      req.session.auth.second_factor = false;
      return SecondFactorGet(vars)(req as any, res as any)
        .then(function () {
          Assert(res.render.calledWith("secondfactor"));
          return BluebirdPromise.resolve();
        });
    });
  });
});
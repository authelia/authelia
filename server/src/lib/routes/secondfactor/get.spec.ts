import SecondFactorGet from "./get";
import { ServerVariablesMockBuilder, ServerVariablesMock }
  from "../../ServerVariablesMockBuilder.spec";
import { ServerVariables } from "../../ServerVariables";
import Sinon = require("sinon");
import ExpressMock = require("../../stubs/express.spec");
import Assert = require("assert");
import Endpoints = require("../../../../../shared/api");
import BluebirdPromise = require("bluebird");

describe("routes/secondfactor/get", function () {
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

  describe("test rendering", function () {
    it("should render second factor page", function () {
      req.session.auth.second_factor = false;
      return SecondFactorGet(vars)(req as any, res as any)
        .then(function () {
          Assert(res.render.calledWith("secondfactor"));
          return BluebirdPromise.resolve();
        });
    });
  });
});
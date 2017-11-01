import Redirect from "../../../src/lib/routes/secondfactor/redirect";
import ExpressMock = require("../../mocks/express");
import { ServerVariablesMockBuilder, ServerVariablesMock }
from "../../mocks/ServerVariablesMockBuilder";
import { ServerVariables } from "../../../src/lib/ServerVariables";
import Assert = require("assert");

describe("test second factor redirect", function() {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let mocks: ServerVariablesMock;
  let vars: ServerVariables;

  beforeEach(function () {
    const s = ServerVariablesMockBuilder.build();
    mocks = s.mocks;
    vars = s.variables;

    req = ExpressMock.RequestMock();
    res = ExpressMock.ResponseMock();
  });

  it("should redirect to default_redirection_url", function() {
    vars.config.default_redirection_url = "http://default_redirection_url";
    Redirect(vars)(req as any, res as any)
    .then(function() {
      Assert(res.json.calledWith({
        redirect: "http://default_redirection_url"
      }));
    });
  });

  it("should redirect to /", function() {
    Redirect(vars)(req as any, res as any)
    .then(function() {
      Assert(res.json.calledWith({
        redirect: "/"
      }));
    });
  });
});
import * as Express from "express";
import Redirect from "./redirect";
import ExpressMock = require("../../stubs/express.spec");
import { ServerVariablesMockBuilder }
from "../../ServerVariablesMockBuilder.spec";
import { ServerVariables } from "../../ServerVariables";
import Assert = require("assert");
import { HEADER_X_TARGET_URL } from "../../constants";

describe("routes/secondfactor/redirect", function() {
  let req: Express.Request;
  let res: ExpressMock.ResponseMock;
  let vars: ServerVariables;

  beforeEach(function () {
    const s = ServerVariablesMockBuilder.build();
    vars = s.variables;

    req = ExpressMock.RequestMock();
    res = ExpressMock.ResponseMock();
    vars.config.session.domain = 'example.com';
  });

  describe('redirect to default url if no target provided', () => {
    it("should redirect to default url", async () => {
      vars.config.default_redirection_url = "https://home.example.com";
      await Redirect(vars)(req, res as any)
      Assert(res.json.calledWith({redirect: "https://home.example.com"}));
    });
  });

  it("should redirect to safe url https://test.example.com/", async () => {
    req.headers[HEADER_X_TARGET_URL] = "https://test.example.com/";
    await Redirect(vars)(req, res as any);
    Assert(res.json.calledWith({redirect: "https://test.example.com/"}));
  });

  it('should not redirect to unsafe target url', async () => {
    vars.config.default_redirection_url = "https://home.example.com";
    req.headers[HEADER_X_TARGET_URL] = "http://test.example.com/";
    await Redirect(vars)(req, res as any);
    Assert(res.status.calledWith(204));
  })
});
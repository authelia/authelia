import Sinon = require("sinon");
import Express = require("express");
import Assert = require("assert");
import Get403 from "../../../../src/lib/routes/error/403/get";
import { ServerVariables } from "../../../../src/lib/ServerVariables";
import { ServerVariablesMockBuilder, ServerVariablesMock }
  from "../../../mocks/ServerVariablesMockBuilder";

describe("Server error 403", function () {
  let vars: ServerVariables;
  let mocks: ServerVariablesMock;
  let req: any;
  let res: any;
  let renderSpy: Sinon.SinonSpy;

  beforeEach(function () {
    const s = ServerVariablesMockBuilder.build();
    vars = s.variables;
    mocks = s.mocks;

    renderSpy = Sinon.spy();
    req = {
      headers: {}
    };
    res = {
      render: renderSpy
    };
  });

  it("should set redirection url to the default redirection url", function () {
    vars.config.default_redirection_url = "http://default-redirection";
    return Get403(vars)(req, res as any)
      .then(function () {
        Assert(renderSpy.calledOnce);
        Assert(renderSpy.calledWithExactly("errors/403", {
          redirection_url: "http://default-redirection"
        }));
      });
  });

  it("should set redirection url to the referer", function () {
    req.headers["referer"] = "http://redirection";
    return Get403(vars)(req, res as any)
      .then(function () {
        Assert(renderSpy.calledOnce);
        Assert(renderSpy.calledWithExactly("errors/403", {
          redirection_url: "http://redirection"
        }));
      });
  });

  it("should render without redirecting the user", function () {
    return Get403(vars)(req, res as any)
      .then(function () {
        Assert(renderSpy.calledOnce);
        Assert(renderSpy.calledWithExactly("errors/403", {
          redirection_url: undefined
        }));
      });
  });
});
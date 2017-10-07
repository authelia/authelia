import Sinon = require("sinon");
import Express = require("express");
import Assert = require("assert");
import Get401 from "../../../../src/lib/routes/error/401/get";

describe("Server error 401", function () {
  it("should render the page", function () {
    const req = {} as Express.Request;
    const res = {
      render: Sinon.stub()
    };

    return Get401(req, res as any)
      .then(function () {
        Assert(res.render.calledOnce);
        Assert(res.render.calledWith("errors/401"));
      });
  });
});
import Sinon = require("sinon");
import Express = require("express");
import Assert = require("assert");
import Get404 from "../../../../../../src/server/lib/routes/error/404/get";

describe("Server error 404", function () {
  it("should render the page", function () {
    const req = {} as Express.Request;
    const res = {
      render: Sinon.stub()
    };

    return Get404(req, res as any)
      .then(function () {
        Assert(res.render.calledOnce);
        Assert(res.render.calledWith("errors/404"));
      });
  });
});
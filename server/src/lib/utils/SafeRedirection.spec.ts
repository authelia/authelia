import Assert = require("assert");
import Sinon = require("sinon");
import { SafeRedirector } from "./SafeRedirection";

describe("web_server/middlewares/SafeRedirection", () => {
  describe("Url is in protected domain", () => {
    before(() => {
      this.redirector = new SafeRedirector("example.com");
      this.res = {redirect: Sinon.stub()};
    });

    it("should redirect to provided url", () => {
      this.redirector.redirectOrElse(this.res,
        "https://mysubdomain.example.com:8080/abc",
        "https://authelia.example.com");
      Assert(this.res.redirect.calledWith("https://mysubdomain.example.com:8080/abc"));
    });

    it("should redirect to default url when wrong domain", () => {
      this.redirector.redirectOrElse(this.res,
        "https://mysubdomain.domain.rtf:8080/abc",
        "https://authelia.example.com");
      Assert(this.res.redirect.calledWith("https://authelia.example.com"));
    });

    it("should redirect to default url when not terminating by domain", () => {
      this.redirector.redirectOrElse(this.res,
        "https://mysubdomain.example.com.rtf:8080/abc",
        "https://authelia.example.com");
      Assert(this.res.redirect.calledWith("https://authelia.example.com"));
    });
  });
});
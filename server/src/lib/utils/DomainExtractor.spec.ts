import { DomainExtractor } from "./DomainExtractor";
import Assert = require("assert");

describe("utils/DomainExtractor", function () {
  describe("test fromUrl", function () {
    it("should return domain from https url", function () {
      const domain = DomainExtractor.fromUrl("https://www.example.com/test/abc");
      Assert.equal(domain, "www.example.com");
    });

    it("should return domain from http url", function () {
      const domain = DomainExtractor.fromUrl("http://www.example.com/test/abc");
      Assert.equal(domain, "www.example.com");
    });

    it("should return domain when url contains port", function () {
      const domain = DomainExtractor.fromUrl("https://www.example.com:8080/test/abc");
      Assert.equal(domain, "www.example.com");
    });
  });
});
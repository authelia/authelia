import { DomainExtractor } from "./DomainExtractor";
import Assert = require("assert");

describe("src/lib/DomainExtractor", function () {
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

    it("should return domain when url contains redirect param", function () {
      const domain0 = DomainExtractor.fromUrl("https://www.example.com:8080/test/abc?rd=https://cool.test.com");
      Assert.equal(domain0, "www.example.com");

      const domain1 = DomainExtractor.fromUrl("https://login.example.com:8080/?rd=https://public.example.com:8080/");
      Assert.equal(domain1, "login.example.com");

      const domain2 = DomainExtractor.fromUrl("https://singlefactor.example.com:8080/secret.html");
      Assert.equal(domain2, "singlefactor.example.com");
    });
  });
});
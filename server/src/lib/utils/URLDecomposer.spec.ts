import { URLDecomposer } from "./URLDecomposer";
import Assert = require("assert");

describe("utils/URLDecomposer", function () {
  describe("test fromUrl", function () {
    it("should return domain from https url", function () {
      const d = URLDecomposer.fromUrl("https://www.example.com/test/abc");
      Assert.equal(d.domain, "www.example.com");
      Assert.equal(d.path, "/test/abc");
    });

    it("should return domain from http url", function () {
      const d = URLDecomposer.fromUrl("http://www.example.com/test/abc");
      Assert.equal(d.domain, "www.example.com");
      Assert.equal(d.path, "/test/abc");
    });

    it("should return domain when url contains port", function () {
      const d = URLDecomposer.fromUrl("https://www.example.com:8080/test/abc");
      Assert.equal(d.domain, "www.example.com");
      Assert.equal(d.path, "/test/abc");
    });

    it("should return default path when no path provided", function () {
      const d = URLDecomposer.fromUrl("https://www.example.com:8080");
      Assert.equal(d.domain, "www.example.com");
      Assert.equal(d.path, "/");
    });

    it("should return default path when provided", function () {
      const d = URLDecomposer.fromUrl("https://www.example.com:8080/");
      Assert.equal(d.domain, "www.example.com");
      Assert.equal(d.path, "/");
    });

    it("should return undefined when does not match", function () {
      const d = URLDecomposer.fromUrl("https:///abc/test");
      Assert.equal(d, undefined);
    });

    it("should return undefined when does not match", function () {
      const d = URLDecomposer.fromUrl("https:///abc/test");
      Assert.equal(d, undefined);
    });
  });
});
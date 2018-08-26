import Assert = require("assert");
import { Sanitizer } from "./Sanitizer";

describe("ldap/InputsSanitizer", function () {
  it("should fail when special characters are used", function () {
    Assert.throws(() => { Sanitizer.sanitize("ab,c"); }, Error);
    Assert.throws(() => { Sanitizer.sanitize("a\\bc"); }, Error);
    Assert.throws(() => { Sanitizer.sanitize("a'bc"); }, Error);
    Assert.throws(() => { Sanitizer.sanitize("a#bc"); }, Error);
    Assert.throws(() => { Sanitizer.sanitize("a+bc"); }, Error);
    Assert.throws(() => { Sanitizer.sanitize("a<bc"); }, Error);
    Assert.throws(() => { Sanitizer.sanitize("a>bc"); }, Error);
    Assert.throws(() => { Sanitizer.sanitize("a;bc"); }, Error);
    Assert.throws(() => { Sanitizer.sanitize("a\"bc"); }, Error);
    Assert.throws(() => { Sanitizer.sanitize("a=bc"); }, Error);
  });

  it("should return original string", function () {
    Assert.equal(Sanitizer.sanitize("abcdef"), "abcdef");
  });

  it("should trim", function () {
    Assert.throws(() => { Sanitizer.sanitize("    abc    "); }, Error);
  });
});

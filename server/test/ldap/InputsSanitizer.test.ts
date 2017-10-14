import Assert = require("assert");
import { InputsSanitizer } from "../../src/lib/ldap/InputsSanitizer";

describe("test InputsSanitizer", function () {
  it("should fail when special characters are used", function () {
    Assert.throws(() => { InputsSanitizer.sanitize("ab,c"); }, Error);
    Assert.throws(() => { InputsSanitizer.sanitize("a\\bc"); }, Error);
    Assert.throws(() => { InputsSanitizer.sanitize("a'bc"); }, Error);
    Assert.throws(() => { InputsSanitizer.sanitize("a#bc"); }, Error);
    Assert.throws(() => { InputsSanitizer.sanitize("a+bc"); }, Error);
    Assert.throws(() => { InputsSanitizer.sanitize("a<bc"); }, Error);
    Assert.throws(() => { InputsSanitizer.sanitize("a>bc"); }, Error);
    Assert.throws(() => { InputsSanitizer.sanitize("a;bc"); }, Error);
    Assert.throws(() => { InputsSanitizer.sanitize("a\"bc"); }, Error);
    Assert.throws(() => { InputsSanitizer.sanitize("a=bc"); }, Error);
  });

  it("should return original string", function () {
    Assert.equal(InputsSanitizer.sanitize("abcdef"), "abcdef");
  });

  it("should trim", function () {
    Assert.throws(() => { InputsSanitizer.sanitize("    abc    "); }, Error);
  });
});

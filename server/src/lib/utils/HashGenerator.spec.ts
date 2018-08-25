import Assert = require("assert");
import { HashGenerator } from "./HashGenerator";

describe("utils/HashGenerator", function () {
  it("should compute correct ssha512 (password)", function () {
    return HashGenerator.ssha512("password", 500000, "jgiCMRyGXzoqpxS3")
      .then(function (hash: string) {
        Assert.equal(hash, "{CRYPT}$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/");
      });
  });

  it("should compute correct ssha512 (test)", function () {
    return HashGenerator.ssha512("test", 500000, "abcdefghijklmnop")
      .then(function (hash: string) {
        Assert.equal(hash, "{CRYPT}$6$rounds=500000$abcdefghijklmnop$sTlNGf0VO/HTQIOXemmaBbV28HUch/qhWOA1/4dsDj6CDQYhUgXbYSPL6gccAsWMr2zD5fFWwhKmPdG.yxphs.");
      });
  });
});
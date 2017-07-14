
import Redis = require("redis");
import Assert = require("assert");

const redisOptions = {
    host: "redis",
    port: 6379
};

describe("test redis is correctly used", function () {
    let redisClient: Redis.RedisClient;

    before(function () {
        redisClient = Redis.createClient(redisOptions);
    });

    it("should have registered at least one session", function (done) {
        redisClient.dbsize(function (err: Error, count: number) {
            Assert.equal(1, count);
            done();
        });
    });
});
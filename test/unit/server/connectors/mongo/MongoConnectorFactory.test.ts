import Assert = require("assert");
import { MongoConnectorFactory } from "../../../../../src/server/lib/connectors/mongo/MongoConnectorFactory";

describe("MongoConnectorFactory", function () {
  describe("create", function () {
    it("should create a connector", function () {
      const factory = new MongoConnectorFactory();
      const connector = factory.create("mongodb://test.url");

      Assert(connector);
    });
  });
});

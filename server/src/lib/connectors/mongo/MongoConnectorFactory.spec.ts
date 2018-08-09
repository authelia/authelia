import Assert = require("assert");
import { MongoConnectorFactory } from "./MongoConnectorFactory";

describe("connectors/mongo/MongoConnectorFactory", function () {
  describe("create", function () {
    it("should create a connector", function () {
      const factory = new MongoConnectorFactory();
      const connector = factory.create("mongodb://test.url");

      Assert(connector);
    });
  });
});

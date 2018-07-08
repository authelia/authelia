import Assert = require("assert");
import Sinon = require("sinon");
import MongoDB = require("mongodb");
import BluebirdPromise = require("bluebird");
import { IMongoClient } from "./IMongoClient";
import { MongoConnector } from "./MongoConnector";

describe("connectors/mongo/MongoConnector", function () {
  let mongoClientConnectStub: Sinon.SinonStub;

  describe("create", function () {
    before(function () {
      mongoClientConnectStub = Sinon.stub(MongoDB.MongoClient, "connect");
    });

    after(function() {
      mongoClientConnectStub.restore();
    });

    it("should create a connector", function () {
      const client = { db: Sinon.mock() };
      mongoClientConnectStub.yields(undefined, client);

      const url = "mongodb://test.url";
      const connector = new MongoConnector(url);
      return connector.connect("database")
        .then(function (client: IMongoClient) {
          Assert(client);
          Assert(mongoClientConnectStub.calledWith(url));
        });
    });

    it("should fail creating a connector", function () {
      mongoClientConnectStub.yields(new Error("Error while creating mongo client"));

      const url = "mongodb://test.url";
      const connector = new MongoConnector(url);
      return connector.connect("database")
        .then(function () { return BluebirdPromise.reject(new Error("It should not be here")); })
        .error(function (client: IMongoClient) {
          Assert(client);
          Assert(mongoClientConnectStub.calledWith(url));
          return BluebirdPromise.resolve();
        });
    });
  });
});

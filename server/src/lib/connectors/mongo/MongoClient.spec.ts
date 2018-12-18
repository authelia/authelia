import Assert = require("assert");
import Bluebird = require("bluebird");
import MongoDB = require("mongodb");
import Sinon = require("sinon");

import { MongoClient } from "./MongoClient";
import { GlobalLoggerStub } from "../../logging/GlobalLoggerStub.spec";
import { MongoStorageConfiguration } from "../../configuration/schema/StorageConfiguration";

describe("connectors/mongo/MongoClient", function () {
  let MongoClientStub: any;
  let mongoClientStub: any;
  let mongoDatabaseStub: any;
  let logger: GlobalLoggerStub = new GlobalLoggerStub();

  const configuration: MongoStorageConfiguration = {
    url: "mongo://url",
    database: "databasename"
  };

  describe("connection", () => {
    before(() => {
      mongoClientStub = {
        db: Sinon.stub()
      };
      mongoDatabaseStub = {
        on: Sinon.stub(),
        collection: Sinon.stub()
      }
      MongoClientStub = Sinon.stub(
        MongoDB.MongoClient, "connect");
      MongoClientStub.yields(
        undefined, mongoClientStub);
      mongoClientStub.db.returns(
        mongoDatabaseStub);
    });

    after(() => {
      MongoClientStub.restore();
    });

    it("should use credentials from configuration", () => {
      configuration.auth = {
        username: "authelia",
        password: "authelia_pass"
      };

      const client = new MongoClient(configuration, logger);
      return client.collection("test")
        .then(() => {
          Assert(MongoClientStub.calledWith("mongo://url", {
            auth: {
              user: "authelia",
              password: "authelia_pass"
            }
          }))
        });
    });
  });

  describe("collection", () => {
    before(function() {
      mongoClientStub = {
        db: Sinon.stub()
      };
      mongoDatabaseStub = {
        on: Sinon.stub(),
        collection: Sinon.stub()
      }
    });

    describe("Connection to mongo is ok", function() {
      before(function () {
        MongoClientStub = Sinon.stub(
          MongoDB.MongoClient, "connect");
        MongoClientStub.yields(
          undefined, mongoClientStub);
        mongoClientStub.db.returns(
          mongoDatabaseStub);
      });
  
      after(function () {
        MongoClientStub.restore();
      });
  
      it("should create a collection", function () {
        const COLLECTION_NAME = "mycollection";
        const client = new MongoClient(configuration, logger);
  
        mongoDatabaseStub.collection.returns("COL");
        return client.collection(COLLECTION_NAME)
          .then((collection) => mongoDatabaseStub.collection.calledWith(COLLECTION_NAME));
      });
    });

    describe("Connection to mongo is broken", function() {
      before(function () {
        MongoClientStub = Sinon.stub(
          MongoDB.MongoClient, "connect");
        MongoClientStub.yields(
          new Error("Failed connection"), undefined);
      });
  
      after(function () {
        MongoClientStub.restore();
      });

      it("should fail creating the collection", function() {
        const COLLECTION_NAME = "mycollection";
        const client = new MongoClient(configuration, logger);
  
        mongoDatabaseStub.collection.returns("COL");
        return client.collection(COLLECTION_NAME)
          .then((collection) => Bluebird.reject(new Error("should not be here.")))
          .catch((err) => Bluebird.resolve());
      });
    })
  });
});

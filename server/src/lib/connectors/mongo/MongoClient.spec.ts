import Assert = require("assert");
import Sinon = require("sinon");
import MongoDB = require("mongodb");
import Bluebird = require("bluebird");
import { MongoClient } from "./MongoClient";
import { GlobalLoggerStub } from "../../logging/GlobalLoggerStub.spec";

describe("connectors/mongo/MongoClient", function () {
  let MongoClientStub: any;
  let mongoClientStub: any;
  let mongoDatabaseStub: any;
  let logger: GlobalLoggerStub = new GlobalLoggerStub();

  describe("collection", function () {
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
        const client = new MongoClient("mongo://url", "databasename", logger);
  
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
        const client = new MongoClient("mongo://url", "databasename", logger);
  
        mongoDatabaseStub.collection.returns("COL");
        return client.collection(COLLECTION_NAME)
          .then((collection) => Bluebird.reject(new Error("should not be here")))
          .error((err) => Bluebird.resolve());
      });
    })
  });
});

import BluebirdPromise = require("bluebird");
import { IClient } from "./IClient";
import { IEmailsRetriever } from "./IEmailsRetriever";
import Sinon = require("sinon");

export class EmailsRetrieverStub implements IEmailsRetriever {
  retrieveStub: Sinon.SinonStub;

  constructor() {
    this.retrieveStub = Sinon.stub();
  }

  retrieve(username: string, client?: IClient): BluebirdPromise<string[]> {
    return this.retrieveStub(username, client);
  }
}
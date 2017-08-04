
import Sinon = require("sinon");
import { INotifier } from "../../../../src/client/lib/INotifier";

export class NotifierStub implements INotifier {
  successStub: Sinon.SinonStub;
  errorStub: Sinon.SinonStub;
  warnStub: Sinon.SinonStub;
  infoStub: Sinon.SinonStub;

  constructor() {
    this.successStub = Sinon.stub();
    this.errorStub = Sinon.stub();
    this.warnStub = Sinon.stub();
    this.infoStub = Sinon.stub();
  }

  success(msg: string) {
    this.successStub();
  }

  error(msg: string) {
    this.errorStub();
  }

  warning(msg: string) {
    this.warnStub();
  }

  info(msg: string) {
    this.infoStub();
  }
}
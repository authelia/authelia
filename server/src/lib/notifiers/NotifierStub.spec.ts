import Sinon = require("sinon");
import BluebirdPromise = require("bluebird");

import { INotifier } from "./INotifier";

export class NotifierStub implements INotifier {
    notifyStub: Sinon.SinonStub;

    constructor() {
        this.notifyStub = Sinon.stub();
    }

    notify(to: string, subject: string, link: string): BluebirdPromise<void> {
        return this.notifyStub(to, subject, link);
    }
}
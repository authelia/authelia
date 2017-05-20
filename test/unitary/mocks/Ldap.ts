
import sinon = require("sinon");

export = function () {
    return {
        bind: sinon.stub(),
        get_emails: sinon.stub(),
        get_groups: sinon.stub()
    };
};

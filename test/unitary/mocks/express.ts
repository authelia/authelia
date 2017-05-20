
import sinon = require("sinon");

export = {
    Response: function () {
        return {
            send: sinon.stub(),
            status: sinon.stub()
        };
    }
};
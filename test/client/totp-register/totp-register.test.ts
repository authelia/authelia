
import sinon = require("sinon");
import assert = require("assert");

import UISelector = require("../../../src/client/totp-register/ui-selector");
import TOTPRegister = require("../../../src/client/totp-register/totp-register");

describe("test totp-register", function() {
    let jqueryMock: any;
    let windowMock: any;
    before(function() {
        jqueryMock = sinon.stub();
        windowMock = {
            QRCode: sinon.spy()
        };
    });

    it("should create qrcode in page", function() {
        const mock = {
            text: sinon.stub(),
            empty: sinon.stub(),
            get: sinon.stub()
        };
        jqueryMock.withArgs(UISelector.QRCODE_ID_SELECTOR).returns(mock);

        TOTPRegister.default(windowMock, jqueryMock);

        assert(mock.text.calledOnce);
        assert(mock.empty.calledOnce);
    });
});
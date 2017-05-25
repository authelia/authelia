
import Endpoints = require("../../../src/server/endpoints");
import BluebirdPromise = require("bluebird");

import UISelectors = require("../../../src/client/firstfactor/UISelectors");
import firstfactor from "../../../src/client/firstfactor/index";
import JQueryMock = require("../mocks/jquery");
import Assert = require("assert");
import sinon = require("sinon");
import jslogger = require("js-logger");

describe("test first factor page", () => {
    it("should validate first factor", () => {
        const jQuery = JQueryMock.JQueryMock();
        const window = {
            location: {
                search: "?redirect=https://example.com",
                href: ""
            },
            document: {},
        };

        const thenSpy = sinon.spy();
        const FirstFactorValidator: any = {
            validate: sinon.stub().returns({ then: thenSpy })
        };

        firstfactor(window as Window, jQuery as any, FirstFactorValidator, jslogger);
        const readyCallback = jQuery.getCall(0).returnValue.ready.getCall(0).args[0];
        readyCallback();

        const onSubmitCallback = jQuery.getCall(1).returnValue.on.getCall(0).args[1];
        jQuery.onCall(2).returns({ val: sinon.stub() });
        jQuery.onCall(3).returns({ val: sinon.stub() });
        jQuery.onCall(4).returns({ val: sinon.stub() });
        jQuery.onCall(5).returns({ val: sinon.stub() });

        onSubmitCallback();

        const successCallback = thenSpy.getCall(0).args[0];
        successCallback();

        Assert.equal(window.location.href, Endpoints.SECOND_FACTOR_GET);
    });

    describe("fail to validate first factor", () => {
        let jQuery: JQueryMock.JQueryMock;
        beforeEach(function () {
            jQuery = JQueryMock.JQueryMock();
            const window = {
                location: {
                    search: "?redirect=https://example.com",
                    href: ""
                },
                document: {},
            };

            const thenSpy = sinon.spy();
            const FirstFactorValidator: any = {
                validate: sinon.stub().returns({ then: thenSpy })
            };

            firstfactor(window as Window, jQuery as any, FirstFactorValidator, jslogger);
            const readyCallback = jQuery.getCall(0).returnValue.ready.getCall(0).args[0];
            readyCallback();

            const onSubmitCallback = jQuery.getCall(1).returnValue.on.getCall(0).args[1];
            jQuery.onCall(2).returns({ val: sinon.stub() });
            jQuery.onCall(3).returns({ val: sinon.stub() });
            jQuery.onCall(4).returns({ val: sinon.stub() });
            jQuery.onCall(5).returns({ val: sinon.stub() });

            onSubmitCallback();

            const failureCallback = thenSpy.getCall(0).args[1];
            failureCallback(new Error("Error when validating first factor"));
        });

        it("should notify the user there is a failure", function () {
            Assert(jQuery.notify.calledOnce);
        });

        it("should reset the password field", function () {
            Assert.equal(jQuery.getCall(4).returnValue.val.getCall(0).args[0], "");
        });
    });
});

import U2FValidator = require("../../src/lib/secondfactor/U2FValidator");
import { INotifier } from "../../src/lib/INotifier";
import JQueryMock = require("../mocks/jquery");
import U2FApiMock = require("../mocks/u2f-api");
import { SignMessage } from "../../../shared/SignMessage";
import BluebirdPromise = require("bluebird");
import Assert = require("assert");
import { NotifierStub } from "../mocks/NotifierStub";

describe("test U2F validation", function () {
    let notifier: INotifier;

    beforeEach(function() {
        notifier = new NotifierStub();
    });

    it("should validate the U2F device", () => {
        const signatureRequest: SignMessage = {
            keyHandle: "keyhandle",
            request: {
                version: "U2F_V2",
                appId: "https://example.com",
                challenge: "challenge"
            }
        };
        const u2fClient = U2FApiMock.U2FApiMock();
        u2fClient.sign.returns(BluebirdPromise.resolve());

        const getPromise = JQueryMock.JQueryDeferredMock();
        getPromise.done.yields(signatureRequest);
        getPromise.done.returns(getPromise);

        const postPromise = JQueryMock.JQueryDeferredMock();
        postPromise.done.yields();
        postPromise.done.returns(postPromise);

        const jqueryMock = JQueryMock.JQueryMock();
        jqueryMock.jquery.get.returns(getPromise);
        jqueryMock.jquery.ajax.returns(postPromise);

        return U2FValidator.validate(jqueryMock.jquery as any, notifier, u2fClient as any);
    });

    it("should fail during initial authentication request", () => {
        const u2fClient = U2FApiMock.U2FApiMock();

        const getPromise = JQueryMock.JQueryDeferredMock();
        getPromise.done.returns(getPromise);
        getPromise.fail.yields(undefined, "Error while issuing authentication request");

        const jqueryMock = JQueryMock.JQueryMock();
        jqueryMock.jquery.get.returns(getPromise);

        return U2FValidator.validate(jqueryMock.jquery as any,  notifier, u2fClient as any)
        .catch(function(err: Error) {
            Assert.equal("Error while issuing authentication request", err.message);
            return BluebirdPromise.resolve();
        });
    });

    it("should fail during device signature", () => {
        const signatureRequest: SignMessage = {
            keyHandle: "keyhandle",
            request: {
                version: "U2F_V2",
                appId: "https://example.com",
                challenge: "challenge"
            }
        };
        const u2fClient = U2FApiMock.U2FApiMock();
        u2fClient.sign.returns(BluebirdPromise.reject(new Error("Device unable to sign")));

        const getPromise = JQueryMock.JQueryDeferredMock();
        getPromise.done.yields(signatureRequest);
        getPromise.done.returns(getPromise);

        const jqueryMock = JQueryMock.JQueryMock();
        jqueryMock.jquery.get.returns(getPromise);

        return U2FValidator.validate(jqueryMock.jquery as any,  notifier, u2fClient as any)
        .catch(function(err: Error) {
            Assert.equal("Device unable to sign", err.message);
            return BluebirdPromise.resolve();
        });
    });

    it("should fail at the end of the authentication request", () => {
        const signatureRequest: SignMessage = {
            keyHandle: "keyhandle",
            request: {
                version: "U2F_V2",
                appId: "https://example.com",
                challenge: "challenge"
            }
        };
        const u2fClient = U2FApiMock.U2FApiMock();
        u2fClient.sign.returns(BluebirdPromise.resolve());

        const getPromise = JQueryMock.JQueryDeferredMock();
        getPromise.done.yields(signatureRequest);
        getPromise.done.returns(getPromise);

        const postPromise = JQueryMock.JQueryDeferredMock();
        postPromise.fail.yields(undefined, "Error while finishing authentication");
        postPromise.done.returns(postPromise);

        const jqueryMock = JQueryMock.JQueryMock();
        jqueryMock.jquery.get.returns(getPromise);
        jqueryMock.jquery.ajax.returns(postPromise);

        return U2FValidator.validate(jqueryMock.jquery as any,  notifier, u2fClient as any)
        .catch(function(err: Error) {
            Assert.equal("Error while finishing authentication", err.message);
            return BluebirdPromise.resolve();
        });
    });
});
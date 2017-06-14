
import TOTPValidator = require("../../../src/client/secondfactor/TOTPValidator");
import JQueryMock = require("../mocks/jquery");
import BluebirdPromise = require("bluebird");
import Assert = require("assert");

describe("test TOTPValidator", function () {
    it("should initiate an identity check successfully", () => {
        const postPromise = JQueryMock.JQueryDeferredMock();
        postPromise.done.yields();
        postPromise.done.returns(postPromise);

        const jqueryMock = JQueryMock.JQueryMock();
        jqueryMock.ajax.returns(postPromise);

        return TOTPValidator.validate("totp_token", jqueryMock as any);
    });

    it("should fail validating TOTP token", () => {
        const errorMessage = "Error while validating TOTP token";

        const postPromise = JQueryMock.JQueryDeferredMock();
        postPromise.fail.yields(undefined, errorMessage);
        postPromise.done.returns(postPromise);

        const jqueryMock = JQueryMock.JQueryMock();
        jqueryMock.ajax.returns(postPromise);

        return TOTPValidator.validate("totp_token", jqueryMock as any)
            .then(function () {
                return BluebirdPromise.reject(new Error("Registration successfully finished while it should have not."));
            }, function (err: Error) {
                Assert.equal(errorMessage, err.message);
                return BluebirdPromise.resolve();
            });
    });
});
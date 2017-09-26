
import FirstFactorValidator = require("../../../../src/client/lib/firstfactor/FirstFactorValidator");
import JQueryMock = require("../mocks/jquery");
import BluebirdPromise = require("bluebird");
import Assert = require("assert");

describe("test FirstFactorValidator", function () {
    it("should validate first factor successfully", () => {
        const postPromise = JQueryMock.JQueryDeferredMock();
        postPromise.done.yields({ redirect: "http://redirect" });
        postPromise.done.returns(postPromise);

        const jqueryMock = JQueryMock.JQueryMock();
        jqueryMock.jquery.ajax.returns(postPromise);

        return FirstFactorValidator.validate("username", "password", "http://redirect", false, jqueryMock.jquery as any);
    });

    function should_fail_first_factor_validation(errorMessage: string) {
        const xhr = {
            status: 401
        };
        const postPromise = JQueryMock.JQueryDeferredMock();
        postPromise.fail.yields(xhr, errorMessage);
        postPromise.done.returns(postPromise);

        const jqueryMock = JQueryMock.JQueryMock();
        jqueryMock.jquery.ajax.returns(postPromise);

        return FirstFactorValidator.validate("username", "password", "http://redirect", false, jqueryMock.jquery as any)
            .then(function () {
                return BluebirdPromise.reject(new Error("First factor validation successfully finished while it should have not."));
            }, function (err: Error) {
                Assert.equal(errorMessage, err.message);
                return BluebirdPromise.resolve();
            });
    }

    describe("should fail first factor validation", () => {
        it("should fail with error", () => {
            return should_fail_first_factor_validation("Authetication failed. Please check your credentials.");
        });
    });
});
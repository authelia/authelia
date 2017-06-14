
import FirstFactorValidator = require("../../../src/client/firstfactor/FirstFactorValidator");
import JQueryMock = require("../mocks/jquery");
import BluebirdPromise = require("bluebird");
import Assert = require("assert");

describe("test FirstFactorValidator", function () {
    it("should validate first factor successfully", () => {
        const postPromise = JQueryMock.JQueryDeferredMock();
        postPromise.done.yields();
        postPromise.done.returns(postPromise);

        const jqueryMock = JQueryMock.JQueryMock();
        jqueryMock.post.returns(postPromise);

        return FirstFactorValidator.validate("username", "password", jqueryMock as any);
    });

    function should_fail_first_factor_validation(statusCode: number, errorMessage: string) {
        const xhr = {
            status: statusCode
        };
        const postPromise = JQueryMock.JQueryDeferredMock();
        postPromise.fail.yields(xhr, errorMessage);
        postPromise.done.returns(postPromise);

        const jqueryMock = JQueryMock.JQueryMock();
        jqueryMock.post.returns(postPromise);

        return FirstFactorValidator.validate("username", "password", jqueryMock as any)
            .then(function () {
                return BluebirdPromise.reject(new Error("First factor validation successfully finished while it should have not."));
            }, function (err: Error) {
                Assert.equal(errorMessage, err.message);
                return BluebirdPromise.resolve();
            });
    }

    describe("should fail first factor validation", () => {
        it("should fail with error 500", () => {
            return should_fail_first_factor_validation(500, "Internal error");
        });

        it("should fail with error 401", () => {
            return should_fail_first_factor_validation(401, "Authetication failed. Please check your credentials");
        });
    });
});
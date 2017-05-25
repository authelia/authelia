
import PasswordResetHandler from "../../../../../src/server/lib/routes/password-reset/identity/PasswordResetHandler";
import LdapClient = require("../../../../../src/server/lib/LdapClient");
import sinon = require("sinon");
import winston = require("winston");
import assert = require("assert");
import BluebirdPromise = require("bluebird");

import ExpressMock = require("../../../mocks/express");
import { LdapClientMock } from "../../../mocks/LdapClient";
import { UserDataStore } from "../../../mocks/UserDataStore";
import ServerVariablesMock = require("../../../mocks/ServerVariablesMock");

describe("test reset password identity check", function () {
    let req: ExpressMock.RequestMock;
    let res: ExpressMock.ResponseMock;
    let user_data_store: UserDataStore;
    let ldap_client: LdapClientMock;
    let configuration: any;

    beforeEach(function () {
        req = {
            query: {
                userid: "user"
            },
            app: {
                get: sinon.stub()
            },
            session: {
                auth_session: {
                    userid: "user",
                    email: "user@example.com",
                    first_factor: true,
                    second_factor: false
                }
            },
            headers: {
                host: "localhost"
            }
        };

        const options = {
            inMemoryOnly: true
        };

        const mocks = ServerVariablesMock.mock(req.app);


        user_data_store = UserDataStore();
        user_data_store.set_u2f_meta.returns(BluebirdPromise.resolve({}));
        user_data_store.get_u2f_meta.returns(BluebirdPromise.resolve({}));
        user_data_store.issue_identity_check_token.returns(BluebirdPromise.resolve({}));
        user_data_store.consume_identity_check_token.returns(BluebirdPromise.resolve({}));
        mocks.userDataStore = user_data_store;


        configuration = {
            ldap: {
                base_dn: "dc=example,dc=com",
                user_name_attribute: "cn"
            }
        };

        mocks.logger = winston;
        mocks.config = configuration;

        ldap_client = LdapClientMock();
        mocks.ldap = ldap_client;

        res = ExpressMock.ResponseMock();
    });

    describe("test reset password identity pre check", () => {
        it("should fail when no userid is provided", function () {
            req.query.userid = undefined;
            const handler = new PasswordResetHandler();
            return handler.preValidationInit(req as any)
                .then(function () { return BluebirdPromise.reject("It should fail"); })
                .catch(function (err: Error) {
                    return BluebirdPromise.resolve();
                });
        });

        it("should fail if ldap fail", function (done) {
            ldap_client.get_emails.returns(BluebirdPromise.reject("Internal error"));
            new PasswordResetHandler().preValidationInit(req as any)
                .catch(function (err: Error) {
                    done();
                });
        });

        it("should perform a search in ldap to find email address", function (done) {
            configuration.ldap.user_name_attribute = "uid";
            ldap_client.get_emails.returns(BluebirdPromise.resolve([]));
            new PasswordResetHandler().preValidationInit(req as any)
                .then(function () {
                    assert.equal("user", ldap_client.get_emails.getCall(0).args[0]);
                    done();
                });
        });

        it("should returns identity when ldap replies", function (done) {
            ldap_client.get_emails.returns(BluebirdPromise.resolve(["test@example.com"]));
            new PasswordResetHandler().preValidationInit(req as any)
                .then(function () {
                    done();
                });
        });
    });
});

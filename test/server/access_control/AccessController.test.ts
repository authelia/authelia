
import assert = require("assert");
import winston = require("winston");
import { AccessController } from "../../../src/server/lib/access_control/AccessController";
import { ACLConfiguration } from "../../../src/types/Configuration";

describe("test access control manager", function () {
    let accessController: AccessController;
    let configuration: ACLConfiguration;

    beforeEach(function () {
        configuration = {
            default: [],
            users: {},
            groups: {}
        };
        accessController = new AccessController(configuration, winston);
    });

    describe("check access control matching", function () {
        beforeEach(function () {
            configuration.default = ["home.example.com", "*.public.example.com"];
            configuration.users = {
                user1: ["user1.example.com", "user1.mail.example.com"]
            };
            configuration.groups = {
                group1: ["secret2.example.com"],
                group2: ["secret.example.com", "secret1.example.com"]
            };
        });

        it("should allow access to secret.example.com", function () {
            assert(accessController.isDomainAllowedForUser("secret.example.com", "user", ["group1", "group2"]));
        });

        it("should deny access to secret3.example.com", function () {
            assert(!accessController.isDomainAllowedForUser("secret3.example.com", "user", ["group1", "group2"]));
        });

        it("should allow access to home.example.com", function () {
            assert(accessController.isDomainAllowedForUser("home.example.com", "user", ["group1", "group2"]));
        });

        it("should allow access to user1.example.com", function () {
            assert(accessController.isDomainAllowedForUser("user1.example.com", "user1", ["group1", "group2"]));
        });

        it("should allow access *.public.example.com", function () {
            assert(accessController.isDomainAllowedForUser("user.public.example.com", "nouser", []));
            assert(accessController.isDomainAllowedForUser("test.public.example.com", "nouser", []));
        });
    });
});

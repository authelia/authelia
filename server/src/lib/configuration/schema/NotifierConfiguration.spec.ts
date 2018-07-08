import Assert = require("assert");
import { NotifierConfiguration, complete } from "./NotifierConfiguration";

describe("configuration/schema/NotifierConfiguration", function() {
  it("should ensure at least one key is provided", function() {
    const configuration: NotifierConfiguration = {};
    const [newConfiguration, error] = complete(configuration);

    Assert.equal(error, "Notifier must have one of the following keys: 'filesystem', 'email' or 'smtp'");
  });

  it("should ensure there is no more than one key", function() {
    const configuration: NotifierConfiguration = {
      smtp: {
        host: "smtp.example.com",
        port: 25,
        secure: false,
        sender: "test@example.com"
      },
      email: {
        username: "test",
        password: "test",
        sender: "test@example.com",
        service: "gmail"
      }
    };
    const [newConfiguration, error] = complete(configuration);

    Assert.equal(error, "Notifier must have one of the following keys: 'filesystem', 'email' or 'smtp'");
  });
});
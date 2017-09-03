
import Assert = require("assert");
import Sinon = require("sinon");
import JQueryMock = require("./mocks/jquery");

import { Notifier } from "../../../src/client/lib/Notifier";

describe("test notifier", function() {
  const SELECTOR = "dummy-selector";
  const MESSAGE = "This is a message";
  let jqueryMock: { jquery: JQueryMock.JQueryMock, element: JQueryMock.JQueryElementsMock };
  let clock: any;

  beforeEach(function() {
    jqueryMock = JQueryMock.JQueryMock();
    clock = Sinon.useFakeTimers();
  });

  afterEach(function() {
    clock.restore();
  });

  function should_fade_in_and_out_on_notification(notificationType: string): void {
    const delayReturn = {
      fadeOut: Sinon.stub()
    };

    jqueryMock.element.fadeIn.yields();

    function onFadedInCallback() {
      Assert(jqueryMock.element.fadeIn.calledOnce);
      Assert(jqueryMock.element.addClass.calledWith(notificationType));
      Assert(!jqueryMock.element.removeClass.calledWith(notificationType));
      clock.tick(10 * 1000);
    }

    function onFadedOutCallback() {
      Assert(jqueryMock.element.removeClass.calledWith(notificationType));
      Assert(jqueryMock.element.fadeOut.calledOnce);
    }

    const notifier = new Notifier(SELECTOR, jqueryMock.jquery as any);

    // Call the method by its name... Bad but allows code reuse.
    (notifier as any)[notificationType](MESSAGE, {
      onFadedIn: onFadedInCallback,
      onFadedOut: onFadedOutCallback
    });

    clock.tick(510);

    Assert(jqueryMock.element.fadeIn.calledOnce);
  }


  it("should fade in and fade out an error message", function() {
    should_fade_in_and_out_on_notification("error");
  });

  it("should fade in and fade out an info message", function() {
    should_fade_in_and_out_on_notification("info");
  });

  it("should fade in and fade out an warning message", function() {
    should_fade_in_and_out_on_notification("warning");
  });

  it("should fade in and fade out an success message", function() {
    should_fade_in_and_out_on_notification("success");
  });
});
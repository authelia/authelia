

import util = require("util");
import { INotifier, Handlers } from "./INotifier";

class NotificationEvent {
  private element: JQuery;
  private message: string;
  private statusType: string;
  private timeoutId: any;

  constructor(element: JQuery, msg: string, statusType: string) {
    this.message = msg;
    this.statusType = statusType;
    this.element = element;
  }

  private clearNotification() {
    this.element.removeClass(this.statusType);
    this.element.html("");
  }

  start(handlers?: Handlers) {
    const that = this;
    const FADE_TIME = 500;
    const html = util.format('<i><img src="/img/notifications/%s.png" alt="status %s"/></i>\
      <span>%s</span>', this.statusType, this.statusType, this.message);
    this.element.html(html);
    this.element.addClass(this.statusType);
    this.element.fadeIn(FADE_TIME, function () {
      handlers.onFadedIn();
    });

    this.timeoutId = setTimeout(function () {
      that.element.fadeOut(FADE_TIME, function () {
        that.clearNotification();
        handlers.onFadedOut();
      });
    }, 4000);
  }

  interrupt() {
    this.clearNotification();
    this.element.hide();
    clearTimeout(this.timeoutId);
  }
}

export class Notifier implements INotifier {
  private element: JQuery;
  private onGoingEvent: NotificationEvent;

  constructor(selector: string, $: JQueryStatic) {
    this.element = $(selector);
    this.onGoingEvent = undefined;
  }

  private displayAndFadeout(msg: string, statusType: string, handlers?: Handlers): void {
    if (this.onGoingEvent)
      this.onGoingEvent.interrupt();

    this.onGoingEvent = new NotificationEvent(this.element, msg, statusType);
    this.onGoingEvent.start();
  }

  success(msg: string, handlers?: Handlers) {
    this.displayAndFadeout(msg, "success", handlers);
  }

  error(msg: string, handlers?: Handlers) {
    this.displayAndFadeout(msg, "error", handlers);
  }

  warning(msg: string, handlers?: Handlers) {
    this.displayAndFadeout(msg, "warning", handlers);
  }

  info(msg: string, handlers?: Handlers) {
    this.displayAndFadeout(msg, "info", handlers);
  }
}
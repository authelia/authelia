

import util = require("util");
import { INotifier, Handlers } from "./INotifier";

export class Notifier implements INotifier {
  private element: JQuery;

  constructor(selector: string, $: JQueryStatic) {
    this.element = $(selector);
  }

  private displayAndFadeout(msg: string, statusType: string, handlers?: Handlers): void {
    const that = this;
    const FADE_TIME = 500;
    const html = util.format('<i><img src="/img/notifications/%s.png" alt="status %s"/></i>\
      <span>%s</span>', statusType, statusType, msg);
    this.element.html(html);
    this.element.addClass(statusType);
    this.element.fadeIn(FADE_TIME, function() {
      handlers.onFadedIn();
    })
    .delay(4000)
    .fadeOut(FADE_TIME, function() {
      that.element.removeClass(statusType);
      handlers.onFadedOut();
    });
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
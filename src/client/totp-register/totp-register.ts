
import jslogger = require("js-logger");
import UISelector = require("./ui-selector");

export default function(window: Window, $: JQueryStatic) {
  jslogger.debug("Creating QRCode from OTPAuth url");
  const qrcode = $(UISelector.QRCODE_ID_SELECTOR);
  const val = qrcode.text();
  qrcode.empty();
  new (window as any).QRCode(qrcode.get(0), val);
}

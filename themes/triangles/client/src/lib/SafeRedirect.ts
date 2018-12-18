import { BelongToDomain } from "../../../shared/BelongToDomain";

export function SafeRedirect(url: string, cb: () => void): void {
  const domain = window.location.hostname.split(".").slice(-2).join(".");
  if (url.startsWith("/") || BelongToDomain(url, domain)) {
    window.location.href = url;
    return;
  }
  cb();
}
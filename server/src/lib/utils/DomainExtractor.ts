export class DomainExtractor {
  static fromUrl(url: string): string {
    if (!url) return "";
    return url.match(/https?:\/\/([^\/:]+).*/)[1];
  }

  static fromHostHeader(host: string): string {
    if (!host) return "";
    return host.split(":")[0];
  }
}
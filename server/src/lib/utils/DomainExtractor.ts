export class DomainExtractor {
  static fromUrl(url: string): string {
    return url.match(/https?:\/\/([^\/:]+).*/)[1];
  }

  static fromHostHeader(host: string): string {
    return host.split(":")[0];
  }
}
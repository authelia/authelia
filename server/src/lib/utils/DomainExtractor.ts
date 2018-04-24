export class DomainExtractor {
  static fromUrl(url: string): string {
    if (!url) return "";
    return url.match(/https?:\/\/([^\/:]+).*/)[1];
  }
}
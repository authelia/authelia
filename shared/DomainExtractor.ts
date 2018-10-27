export class DomainExtractor {
  static fromUrl(url: string): string {
    if (!url) return;
    const matches = url.match(/(https?:\/\/)?([a-zA-Z0-9_.-]+).*/);

    if (matches.length > 2) {
      return matches[2];
    }
    return;
  }
}
// TODO: replace this decompose by third party library.
export class URLDecomposer {
  static fromUrl(url: string): {domain: string, path: string} {
    if (!url) return;
    const match = url.match(/https?:\/\/([a-z0-9_.-]+)(:[0-9]+)?(.*)/);

    if (!match) return;

    if (match[1] && !match[3]) {
      return {domain: match[1], path: "/"};
    } else if (match[1] && match[3]) {
      return {domain: match[1], path: match[3]};
    }
    return;
  }
}
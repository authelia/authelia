import { DomainExtractor } from "./DomainExtractor";

export function BelongToDomain(url: string, domain: string): boolean {
  const urlDomain =  DomainExtractor.fromUrl(url);
  if (!urlDomain) return false;
  const idx = urlDomain.indexOf(domain);
  return idx + domain.length == urlDomain.length;
}
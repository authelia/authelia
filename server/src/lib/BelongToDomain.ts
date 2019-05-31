import { DomainExtractor } from "./DomainExtractor";
import { IRequestLogger } from "./logging/IRequestLogger";
import express = require("express");

export function BelongToDomain(url: string, domain: string, logger: IRequestLogger, req: express.Request): boolean {
  const urlDomain =  DomainExtractor.fromUrl(url);
  if (!urlDomain) {
    logger.debug(req, "Unable to extract domain from url %s the url doesn't parse correctly.", url);
    return false;
  }
  logger.debug(req,  "Extracted domain %s from url %s", urlDomain, url);
  const idx = urlDomain.indexOf(domain);
  logger.debug(req,  "Found protected domain: %s in url extracted domain: %s at index: %s",
      domain, urlDomain, idx);
  logger.debug(req, "protected domain size: %s url extracted domain size: %s", domain.length, urlDomain.length);
  logger.debug(req, "domain match url extracted: %s", idx + domain.length == urlDomain.length);
  return idx + domain.length == urlDomain.length;
}
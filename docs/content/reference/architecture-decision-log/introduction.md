---
title: "Architecture Decision Log"
description: "The Authelia Architecture Decision Log"
summary: "This section contains Architecture Decision Records for Authelia."
date: 2024-06-16T09:32:59+10:00
draft: false
images: []
weight: 1000
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The architecture decision log (ADL) is a collection of architecture decision records (ADRs). This is one of the various
methods the Authelia project uses to make and record decisions. It is a semi-formal process which records a specific
scenario or issue and the various implementation options available and their various drawbacks and advantages. It may at
some point have a formal decision that's been made and the specific rational for that decision, as well as the rational
why other implementation was not chosen.

There may or may not be a linked public discussion on a specific ADR, there may be incremental revisions as new
information comes to light, and various other changes as time progresses.

The intent behind the Architecture Decision Log has primary and secondary elements.

Primarily:

1. Have formal discussions on various elements of the design process and record formal decisions
2. Keep a log of historical choices so that we can refer back to them in the event we forget the nuances of a complex
   decision

Secondarily:

1. Reference the specific ADR's in other areas of the documentation
2. Clearly and efficiently communicate a specific decision in the event of a feature request or bug report

Each ADR has several key sections.

|      Name       |                             Description                              |
|:---------------:|:--------------------------------------------------------------------:|
|     Status      |                    The current status of the ADR                     |
|     Context     | The context surrounding why the ADR is necessary and what it's about |
| Proposed Design |           The proposed design or designs for consideration           |
|    Decision     |               The decision made as a result of the ADR               |
|  Consequences   |        The potential undesirable outcomes related to the ADR         |
|  Related ADRs   |          The other ADRs that are related to a specific ADR           |

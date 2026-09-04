---
title: "Post-Quantum Cryptography"
description: "An overview of the Post-Quantum Cryptography security Authelia implements."
summary: "An overview of the Post-Quantum Cryptography security Authelia implements."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 420
toc: true
aliases: []
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

With the rapid improvements to Quantum Computing; risks to traditional cryptography have been accelerating. With
algorithms like Shor's algorithm and Grover's algorithm both asymmetric and symmetric encryption respectively are able
to be broken in much more efficient ways.

This has lead to research into cryptographic methods that are based on very different mathematical foundations which are
often based on high-dimensional lattices with randomized noise. The known Quantum Computing algebraic shortcuts that can
achieve these efficient breakage are not at all usable against these lattice problems at this time.

These developments have also lead many countries to require adoption of these standards either in government agencies or
in parties that deal with these agencies. The deadline for these requirements is fast approaching.

As of 4.39.21 released on September 3 2026, Authelia implements several Post-Quantum Cryptographic measures to ensure
we're not only not left behind but also ahead of the curve.

## WebAuthn

As of 4.39.21 released on September 3 2026, supports and requests ML-DSA-44, ML-DSA-65, ML-DSA-87 credentials from
authenticators.

## OpenID Connect 1.0 and OAuth 2.0

As of 4.39.21 released on September 3 2026, supports ML-DSA-44, ML-DSA-65, ML-DSA-87 JOSE signatures for both response
objects and as request objects of almost all types.

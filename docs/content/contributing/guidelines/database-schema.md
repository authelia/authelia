---
title: "Database Schema"
description: "Authelia Development Database Schema Guidelines"
summary: "This section covers the database schema guidelines we use for development."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 320
toc: true
aliases: []
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Table Names

1. Should match in every database implementation.
2. Should be all lower case.
3. Should use singular form (i.e. not plural).
4. Should use the underscore character (`_`) between words.
5. Should only contain alphanumeric characters and the underscore character (`_`).
   1. The underscore character (`_`):
      1. Should always be used between words.
      2. Should only be used:
         1. Between words.
         2. As a prefix for temporary tables.
   2. Should start and end with only an alphabetic character, excluding specific exceptions mentioned elsewhere with
      prefix and suffix terminology.

## Column Names

1. Should match in every database implementation.
2. Should be all lower case.
3. Should only contain alphanumeric characters and the underscore character (`_`).
   1. The underscore character (`_`):
      1. Should always be used between words.
      2. Should only be used between words.
   2. Should only start and end with an alphabetic character.

## Key Names

### Foreign Keys

Format: `<table_name>_<column_name>_fkey`

Where:

- The table name is the name of the table the foreign key exists on.
- The column name is the name of the column the foreign key is for.

### Unique Keys

Format: `<table_name>_<key_name>_key`

Where:

- The table name is the name of the table the unique key is on.
- The key name is the name to describe this key. This can also be the column name it exists on.

### Primary Keys

Most database engines don't allow customizing the primary key names. As such the primary key should not be explicitly
set except to change it back to the default format.

---
title: "Database Schema"
description: "Authelia Development Database Schema Guidelines"
lead: "This section covers the database schema guidelines we use for development."
date: 2022-11-09T09:20:18+11:00
draft: false
images: []
menu:
  contributing:
    parent: "guidelines"
weight: 320
toc: true
aliases: []
---

## Table Names

1. Should match in every database implementation.
2. Should be all lower case.
3. Should use singular form (i.e. not plural).
4. Should use underscores between words (`_`).

## Column Names

1. Should match in every database implementation.
2. Should be all lower case.
3. Should use underscores between words (`_`).

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

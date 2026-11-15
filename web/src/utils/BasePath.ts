// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { getEmbeddedVariable } from "@utils/Configuration";

export function getBasePath() {
    return getEmbeddedVariable("basepath");
}

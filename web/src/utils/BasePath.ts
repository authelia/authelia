// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { getEmbeddedVariable } from "@utils/Configuration";

export function getBasePath() {
    return getEmbeddedVariable("basepath");
}

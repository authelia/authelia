// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { ServiceError } from "@services/ServiceError";

it("carries the status and code", () => {
    const err = new ServiceError("Failed POST to /path. Code: 200. Message: bad", 200, "password_policy");

    expect(err).toBeInstanceOf(Error);
    expect(err.name).toBe("ServiceError");
    expect(err.message).toBe("Failed POST to /path. Code: 200. Message: bad");
    expect(err.status).toBe(200);
    expect(err.code).toBe("password_policy");
});

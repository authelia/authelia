// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { beforeEach, describe, expect, it, vi } from "vitest";

import { FlowContinuePath } from "@services/Api";
import { PostWithOptionalResponse } from "@services/Client";
import { postFlowContinue } from "@services/Flow";

vi.mock("@services/Client");

describe("postFlowContinue", () => {
    beforeEach(() => {
        vi.resetAllMocks();
    });

    it("should post the flow parameters and return the redirect", async () => {
        vi.mocked(PostWithOptionalResponse).mockResolvedValue({
            redirect: "https://auth.example.com/api/oidc/authorization",
        });

        const result = await postFlowContinue("abc-123", "openid_connect", undefined, undefined);

        expect(PostWithOptionalResponse).toHaveBeenCalledWith(FlowContinuePath, {
            flow: "openid_connect",
            flowID: "abc-123",
            subflow: undefined,
            userCode: undefined,
        });

        expect(result).toEqual({ redirect: "https://auth.example.com/api/oidc/authorization" });
    });

    it("should return an empty response when the server replies without a body", async () => {
        vi.mocked(PostWithOptionalResponse).mockResolvedValue(undefined);

        const result = await postFlowContinue("abc-123", "openid_connect", undefined, undefined);

        expect(result).toEqual({});
    });
});

// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import i18n from "i18next";

import { Post } from "@services/Client";
import { postExternalIdentityStart } from "@services/ExternalIdentity";

vi.mock("@services/Client", () => ({
    DeleteWithOptionalResponse: vi.fn(),
    Get: vi.fn(),
    Post: vi.fn(),
    PutWithOptionalResponse: vi.fn(),
}));

afterEach(async () => {
    await i18n.changeLanguage("en");
});

it("sends the language the user chose with the start request", async () => {
    vi.mocked(Post).mockResolvedValue({ authorization_url: "https://app.plex.tv/auth" });

    await i18n.changeLanguage("fr-CA");

    const response = await postExternalIdentityStart("plex", {
        keepMeLoggedIn: true,
        targetURL: "https://app.example.com",
    });

    expect(response).toEqual({ authorization_url: "https://app.plex.tv/auth" });
    expect(Post).toHaveBeenCalledWith(
        expect.stringMatching(/\/api\/firstfactor\/external-identity\/plex$/),
        { keepMeLoggedIn: true, language: "fr-CA", targetURL: "https://app.example.com" },
        undefined,
    );
});

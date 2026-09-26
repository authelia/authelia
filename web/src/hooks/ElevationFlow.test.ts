// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook, waitFor } from "@testing-library/react";

import { isElevated, useElevationFlow } from "@hooks/ElevationFlow";
import { getUserSessionElevation } from "@services/UserSessionElevation";

vi.mock("@services/UserSessionElevation", () => ({
    getUserSessionElevation: vi.fn(),
}));

const getElevationMock = vi.mocked(getUserSessionElevation);

const elevated = { elevated: true, skip_second_factor: false } as any;
const notElevated = { elevated: false, skip_second_factor: false } as any;
const skipSecondFactor = { elevated: false, skip_second_factor: true } as any;
const reauthenticationRequired = {
    elevated: true,
    reauthentication_methods: ["password"],
    require_reauthentication: true,
    skip_second_factor: false,
} as any;

function setup(onRefreshError?: (error: unknown) => void) {
    const onElevated = vi.fn();
    const onCancelled = vi.fn();

    const hook = renderHook(() => useElevationFlow<"delete" | "register">({ onCancelled, onElevated, onRefreshError }));

    return { hook, onCancelled, onElevated };
}

function deferred<T>() {
    let resolve!: (value: T) => void;
    let reject!: (reason?: unknown) => void;
    const promise = new Promise<T>((res, rej) => {
        resolve = res;
        reject = rej;
    });

    return { promise, reject, resolve };
}

async function startAndPassReauthentication(hook: ReturnType<typeof setup>["hook"]) {
    act(() => hook.result.current.start("register"));
    await waitFor(() => expect(hook.result.current.reauthenticationDialogProps.elevation).toBeDefined());
    act(() => hook.result.current.reauthenticationDialogProps.handleClosed(true, false));
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.spyOn(console, "error").mockImplementation(() => {});
    getElevationMock.mockResolvedValue(elevated);
});

afterEach(() => {
    vi.restoreAllMocks();
});

describe("isElevated", () => {
    it("treats elevated and skip_second_factor as elevated", () => {
        expect(isElevated(elevated)).toBe(true);
        expect(isElevated(skipSecondFactor)).toBe(true);
        expect(isElevated(notElevated)).toBe(false);
        expect(isElevated(undefined)).toBe(false);
    });
});

describe("start", () => {
    it("fetches the elevation and opens the reauthentication dialog", async () => {
        const { hook } = setup();

        act(() => hook.result.current.start("register"));

        expect(hook.result.current.pending).toBe("register");
        expect(hook.result.current.reauthenticationDialogProps.opening).toBe(true);
        expect(hook.result.current.secondFactorDialogProps.opening).toBe(false);
        await waitFor(() => expect(hook.result.current.reauthenticationDialogProps.elevation).toEqual(elevated));
    });

    it("logs elevation lookup failures", async () => {
        getElevationMock.mockRejectedValue(new Error("boom"));
        const { hook } = setup();

        act(() => hook.result.current.start("register"));

        await waitFor(() => expect(console.error).toHaveBeenCalled());
    });
});

describe("reauthentication dialog", () => {
    it("clears the opening flag once it opens", () => {
        const { hook } = setup();

        act(() => hook.result.current.start("register"));
        act(() => hook.result.current.reauthenticationDialogProps.handleOpened());

        expect(hook.result.current.reauthenticationDialogProps.opening).toBe(false);
    });

    it("cancels the flow", async () => {
        const { hook, onCancelled } = setup();

        act(() => hook.result.current.start("register"));
        act(() => hook.result.current.reauthenticationDialogProps.handleClosed(false, false));

        expect(onCancelled).toHaveBeenCalledOnce();
        expect(hook.result.current.pending).toBeUndefined();
    });

    it("opens the second factor dialog when unchanged", async () => {
        const { hook } = setup();

        await startAndPassReauthentication(hook);

        expect(hook.result.current.secondFactorDialogProps.opening).toBe(true);
        expect(getElevationMock).toHaveBeenCalledTimes(1);
    });

    it("re-checks the elevation before the second factor dialog when changed", async () => {
        const { hook } = setup();

        act(() => hook.result.current.start("register"));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(1));

        act(() => hook.result.current.reauthenticationDialogProps.handleClosed(true, true));

        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(2));
        await waitFor(() => expect(hook.result.current.secondFactorDialogProps.opening).toBe(true));
    });

    it("reports a failed re-check without advancing", async () => {
        const onRefreshError = vi.fn();
        const { hook } = setup(onRefreshError);

        act(() => hook.result.current.start("register"));
        await waitFor(() => expect(getElevationMock).toHaveBeenCalledTimes(1));

        getElevationMock.mockRejectedValueOnce(new Error("boom"));
        act(() => hook.result.current.reauthenticationDialogProps.handleClosed(true, true));

        await waitFor(() => expect(onRefreshError).toHaveBeenCalledOnce());
        expect(hook.result.current.secondFactorDialogProps.opening).toBe(false);
    });
});

describe("second factor dialog", () => {
    it("clears the opening flag once it opens", async () => {
        const { hook } = setup();

        await startAndPassReauthentication(hook);
        act(() => hook.result.current.secondFactorDialogProps.handleOpened());

        expect(hook.result.current.secondFactorDialogProps.opening).toBe(false);
    });

    it("cancels the flow", async () => {
        const { hook, onCancelled } = setup();

        await startAndPassReauthentication(hook);
        act(() => hook.result.current.secondFactorDialogProps.handleClosed(false, false));

        expect(onCancelled).toHaveBeenCalledOnce();
    });

    it("completes when already elevated", async () => {
        const { hook, onElevated } = setup();

        await startAndPassReauthentication(hook);
        act(() => hook.result.current.secondFactorDialogProps.handleClosed(true, false));

        await waitFor(() => expect(onElevated).toHaveBeenCalledWith("register"));
        expect(hook.result.current.pending).toBeUndefined();
    });

    it("treats skip_second_factor as elevated", async () => {
        getElevationMock.mockResolvedValue(skipSecondFactor);
        const { hook, onElevated } = setup();

        await startAndPassReauthentication(hook);
        act(() => hook.result.current.secondFactorDialogProps.handleClosed(true, false));

        await waitFor(() => expect(onElevated).toHaveBeenCalledWith("register"));
    });

    it("re-checks the reauthentication requirement before completing", async () => {
        const { hook, onElevated } = setup();

        await startAndPassReauthentication(hook);
        expect(getElevationMock).toHaveBeenCalledTimes(1);

        act(() => hook.result.current.secondFactorDialogProps.handleClosed(true, false));

        await waitFor(() => expect(onElevated).toHaveBeenCalledWith("register"));
        expect(getElevationMock).toHaveBeenCalledTimes(2);
    });

    it("returns to reauthentication when the requirement lapsed before completing", async () => {
        const { hook, onElevated } = setup();

        await startAndPassReauthentication(hook);

        getElevationMock.mockResolvedValueOnce(reauthenticationRequired);
        act(() => hook.result.current.secondFactorDialogProps.handleClosed(true, false));

        await waitFor(() => expect(hook.result.current.reauthenticationDialogProps.opening).toBe(true));
        expect(onElevated).not.toHaveBeenCalled();
        expect(hook.result.current.pending).toBe("register");
        expect(hook.result.current.reauthenticationDialogProps.elevation).toEqual(reauthenticationRequired);
    });

    it("returns to reauthentication when the requirement lapsed after a change", async () => {
        getElevationMock
            .mockResolvedValueOnce(notElevated)
            .mockResolvedValueOnce(elevated)
            .mockResolvedValueOnce(reauthenticationRequired);
        const { hook, onElevated } = setup();

        await startAndPassReauthentication(hook);
        act(() => hook.result.current.secondFactorDialogProps.handleClosed(true, true));

        await waitFor(() => expect(hook.result.current.reauthenticationDialogProps.opening).toBe(true));
        expect(getElevationMock).toHaveBeenCalledTimes(3);
        expect(onElevated).not.toHaveBeenCalled();
    });

    it("reports a failed completion re-check without completing", async () => {
        const onRefreshError = vi.fn();
        const { hook, onElevated } = setup(onRefreshError);

        await startAndPassReauthentication(hook);

        getElevationMock.mockRejectedValueOnce(new Error("boom"));
        act(() => hook.result.current.secondFactorDialogProps.handleClosed(true, false));

        await waitFor(() => expect(onRefreshError).toHaveBeenCalledOnce());
        expect(onElevated).not.toHaveBeenCalled();
    });

    it("asks for identity verification when not elevated", async () => {
        getElevationMock.mockResolvedValue(notElevated);
        const { hook, onElevated } = setup();

        await startAndPassReauthentication(hook);
        act(() => hook.result.current.secondFactorDialogProps.handleClosed(true, false));

        expect(onElevated).not.toHaveBeenCalled();
        expect(hook.result.current.identityVerificationDialogProps.opening).toBe(true);
    });

    it("re-checks the elevation when changed", async () => {
        getElevationMock.mockResolvedValueOnce(notElevated).mockResolvedValueOnce(elevated);
        const { hook, onElevated } = setup();

        await startAndPassReauthentication(hook);
        act(() => hook.result.current.secondFactorDialogProps.handleClosed(true, true));

        await waitFor(() => expect(onElevated).toHaveBeenCalledWith("register"));
        expect(getElevationMock).toHaveBeenCalledTimes(3);
    });

    it("asks for identity verification when the re-checked elevation is insufficient", async () => {
        getElevationMock.mockResolvedValue(notElevated);
        const { hook } = setup();

        await startAndPassReauthentication(hook);
        act(() => hook.result.current.secondFactorDialogProps.handleClosed(true, true));

        await waitFor(() => expect(hook.result.current.identityVerificationDialogProps.opening).toBe(true));
    });

    it("reports a failed re-check without advancing", async () => {
        const onRefreshError = vi.fn();
        const { hook, onElevated } = setup(onRefreshError);

        await startAndPassReauthentication(hook);

        getElevationMock.mockRejectedValueOnce(new Error("boom"));
        act(() => hook.result.current.secondFactorDialogProps.handleClosed(true, true));

        await waitFor(() => expect(onRefreshError).toHaveBeenCalledOnce());
        expect(onElevated).not.toHaveBeenCalled();
        expect(hook.result.current.identityVerificationDialogProps.opening).toBe(false);
    });
});

describe("identity verification dialog", () => {
    async function reachIdentityVerification(hook: ReturnType<typeof setup>["hook"]) {
        await startAndPassReauthentication(hook);
        act(() => hook.result.current.secondFactorDialogProps.handleClosed(true, false));
    }

    beforeEach(() => {
        getElevationMock.mockResolvedValue(notElevated);
    });

    it("clears the opening flag once it opens", async () => {
        const { hook } = setup();

        await reachIdentityVerification(hook);
        act(() => hook.result.current.identityVerificationDialogProps.handleOpened());

        expect(hook.result.current.identityVerificationDialogProps.opening).toBe(false);
    });

    it("completes once identity is verified", async () => {
        const { hook, onElevated } = setup();

        await reachIdentityVerification(hook);
        act(() => hook.result.current.identityVerificationDialogProps.handleClosed(true));

        await waitFor(() => expect(onElevated).toHaveBeenCalledWith("register"));
        expect(getElevationMock).toHaveBeenCalledTimes(2);
    });

    it("returns to reauthentication when the requirement lapsed during identity verification", async () => {
        const { hook, onElevated } = setup();

        await reachIdentityVerification(hook);

        getElevationMock.mockResolvedValueOnce(reauthenticationRequired);
        act(() => hook.result.current.identityVerificationDialogProps.handleClosed(true));

        await waitFor(() => expect(hook.result.current.reauthenticationDialogProps.opening).toBe(true));
        expect(onElevated).not.toHaveBeenCalled();
        expect(hook.result.current.identityVerificationDialogProps.opening).toBe(false);
        expect(hook.result.current.pending).toBe("register");
    });

    it("completes when the re-checked reauthentication requirement is satisfied", async () => {
        const { hook, onElevated } = setup();

        await reachIdentityVerification(hook);

        getElevationMock.mockResolvedValueOnce({ ...notElevated, require_reauthentication: false });
        act(() => hook.result.current.identityVerificationDialogProps.handleClosed(true));

        await waitFor(() => expect(onElevated).toHaveBeenCalledWith("register"));
        expect(hook.result.current.reauthenticationDialogProps.opening).toBe(false);
    });

    it("reports a failed completion re-check without completing", async () => {
        const onRefreshError = vi.fn();
        const { hook, onElevated } = setup(onRefreshError);

        await reachIdentityVerification(hook);

        getElevationMock.mockRejectedValueOnce(new Error("boom"));
        act(() => hook.result.current.identityVerificationDialogProps.handleClosed(true));

        await waitFor(() => expect(onRefreshError).toHaveBeenCalledOnce());
        expect(onElevated).not.toHaveBeenCalled();
    });

    it("cancels the flow", async () => {
        const { hook, onCancelled, onElevated } = setup();

        await reachIdentityVerification(hook);
        act(() => hook.result.current.identityVerificationDialogProps.handleClosed(false));

        expect(onCancelled).toHaveBeenCalledOnce();
        expect(onElevated).not.toHaveBeenCalled();
    });
});

describe("reset", () => {
    it("clears the pending action and every opening flag", () => {
        getElevationMock.mockReturnValue(new Promise(() => {}));

        const { hook } = setup();

        act(() => hook.result.current.start("delete"));
        act(() => hook.result.current.reset());

        expect(hook.result.current.pending).toBeUndefined();
        expect(hook.result.current.reauthenticationDialogProps.opening).toBe(false);
        expect(hook.result.current.reauthenticationDialogProps.elevation).toBeUndefined();
    });
});

describe("stale fetches", () => {
    it("does not repopulate the elevation when a fetch resolves after reset", async () => {
        const fetch = deferred<any>();
        getElevationMock.mockReturnValueOnce(fetch.promise);

        const { hook } = setup();

        act(() => hook.result.current.start("register"));
        act(() => hook.result.current.reset());

        await act(async () => {
            fetch.resolve(elevated);
            await fetch.promise;
        });

        expect(hook.result.current.reauthenticationDialogProps.elevation).toBeUndefined();
        expect(hook.result.current.pending).toBeUndefined();
    });

    it("ignores a stale fetch from a previous start", async () => {
        const first = deferred<any>();
        const second = deferred<any>();
        getElevationMock.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);

        const { hook } = setup();

        act(() => hook.result.current.start("register"));
        act(() => hook.result.current.start("delete"));

        await act(async () => {
            second.resolve(elevated);
            await second.promise;
        });

        await act(async () => {
            first.resolve(notElevated);
            await first.promise;
        });

        expect(hook.result.current.reauthenticationDialogProps.elevation).toEqual(elevated);
    });

    it("does not advance when a post-change re-fetch resolves after cancel", async () => {
        const { hook, onElevated } = setup();

        await startAndPassReauthentication(hook);

        const refetch = deferred<any>();
        getElevationMock.mockReturnValueOnce(refetch.promise);

        act(() => hook.result.current.secondFactorDialogProps.handleClosed(true, true));
        act(() => hook.result.current.reset());

        await act(async () => {
            refetch.resolve(elevated);
            await refetch.promise;
        });

        expect(onElevated).not.toHaveBeenCalled();
        expect(hook.result.current.identityVerificationDialogProps.opening).toBe(false);
    });

    it("does not complete when the completion re-check resolves after reset", async () => {
        const { hook, onElevated } = setup();

        await startAndPassReauthentication(hook);

        const recheck = deferred<any>();
        getElevationMock.mockReturnValueOnce(recheck.promise);

        act(() => hook.result.current.secondFactorDialogProps.handleClosed(true, false));
        act(() => hook.result.current.reset());

        await act(async () => {
            recheck.resolve(elevated);
            await recheck.promise;
        });

        expect(onElevated).not.toHaveBeenCalled();
        expect(hook.result.current.reauthenticationDialogProps.opening).toBe(false);
        expect(hook.result.current.reauthenticationDialogProps.elevation).toBeUndefined();
    });
});

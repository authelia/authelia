// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook } from "@testing-library/react";

import { useRememberMePrompt } from "@hooks/RememberMe";

it("starts closed", () => {
    const { result } = renderHook(() => useRememberMePrompt(true));

    expect(result.current.dialogProps.open).toBe(false);
});

it("resolves false without opening when the feature is disabled", async () => {
    const { result } = renderHook(() => useRememberMePrompt(false));

    let resolved: boolean | null | undefined;

    await act(async () => {
        resolved = await result.current.prompt();
    });

    expect(resolved).toBe(false);
    expect(result.current.dialogProps.open).toBe(false);
});

it("opens and resolves with the choice the user makes", async () => {
    const { result } = renderHook(() => useRememberMePrompt(true));

    let resolved: boolean | null | undefined;

    act(() => {
        result.current.prompt().then((value) => (resolved = value));
    });

    expect(result.current.dialogProps.open).toBe(true);
    expect(resolved).toBeUndefined();

    await act(async () => result.current.dialogProps.onChoice(true));

    expect(resolved).toBe(true);
    expect(result.current.dialogProps.open).toBe(false);
});

it("resolves false when the user declines", async () => {
    const { result } = renderHook(() => useRememberMePrompt(true));

    let resolved: boolean | null | undefined;

    act(() => {
        result.current.prompt().then((value) => (resolved = value));
    });

    await act(async () => result.current.dialogProps.onChoice(false));

    expect(resolved).toBe(false);
});

it("cancels an outstanding prompt that a new one supersedes", async () => {
    const { result } = renderHook(() => useRememberMePrompt(true));

    let first: boolean | null | undefined;
    let second: boolean | null | undefined;

    act(() => {
        result.current.prompt().then((value) => (first = value));
    });

    await act(async () => {
        result.current.prompt().then((value) => (second = value));
    });

    expect(first).toBeNull();
    expect(second).toBeUndefined();
    expect(result.current.dialogProps.open).toBe(true);

    await act(async () => result.current.dialogProps.onChoice(true));

    expect(second).toBe(true);
});

it("resolves null when the user cancels", async () => {
    const { result } = renderHook(() => useRememberMePrompt(true));

    let resolved: boolean | null | undefined;

    act(() => {
        result.current.prompt().then((value) => (resolved = value));
    });

    await act(async () => result.current.dialogProps.onCancel());

    expect(resolved).toBeNull();
    expect(result.current.dialogProps.open).toBe(false);
});

it("resolves null when unmounted with a prompt outstanding", async () => {
    const { result, unmount } = renderHook(() => useRememberMePrompt(true));

    let resolved: boolean | null | undefined;

    act(() => {
        result.current.prompt().then((value) => (resolved = value));
    });

    unmount();

    await act(async () => {});

    expect(resolved).toBeNull();
});

it("ignores a choice made while no prompt is outstanding", async () => {
    const { result } = renderHook(() => useRememberMePrompt(true));

    await act(async () => result.current.dialogProps.onChoice(true));

    expect(result.current.dialogProps.open).toBe(false);
});

it("keeps a stable prompt and choice handler across renders", () => {
    const { rerender, result } = renderHook(() => useRememberMePrompt(true));

    const { prompt } = result.current;
    const { onCancel, onChoice } = result.current.dialogProps;

    rerender();

    expect(result.current.prompt).toBe(prompt);
    expect(result.current.dialogProps.onChoice).toBe(onChoice);
    expect(result.current.dialogProps.onCancel).toBe(onCancel);
});

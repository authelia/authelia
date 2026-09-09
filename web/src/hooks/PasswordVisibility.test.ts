import { act, renderHook } from "@testing-library/react";

import { usePasswordVisibility } from "@hooks/PasswordVisibility";

it("starts hidden", () => {
    const { result } = renderHook(() => usePasswordVisibility());

    expect(result.current.showPassword).toBe(false);
});

it("shows the password when toggled and hides it when toggled again", () => {
    const { result } = renderHook(() => usePasswordVisibility());

    act(() => result.current.toggleProps.onClick());
    expect(result.current.showPassword).toBe(true);

    act(() => result.current.toggleProps.onClick());
    expect(result.current.showPassword).toBe(false);
});

it("reflects the visibility state via aria-pressed", () => {
    const { result } = renderHook(() => usePasswordVisibility());

    expect(result.current.toggleProps["aria-pressed"]).toBe(false);

    act(() => result.current.toggleProps.onClick());

    expect(result.current.toggleProps["aria-pressed"]).toBe(true);
});

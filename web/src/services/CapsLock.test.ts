import { IsCapsLockModified } from "@services/CapsLock";

it("returns null for key length not 1", () => {
    const event = { altKey: false, ctrlKey: false, getModifierState: vi.fn(), key: "Enter", metaKey: false } as any;
    expect(IsCapsLockModified(event)).toBeNull();
});

it("returns null for ctrl key", () => {
    const event = { altKey: false, ctrlKey: true, getModifierState: vi.fn(), key: "a", metaKey: false } as any;
    expect(IsCapsLockModified(event)).toBeNull();
});

it("returns null for alt key", () => {
    const event = { altKey: true, ctrlKey: false, getModifierState: vi.fn(), key: "a", metaKey: false } as any;
    expect(IsCapsLockModified(event)).toBeNull();
});

it("returns null for meta key", () => {
    const event = { altKey: false, ctrlKey: false, getModifierState: vi.fn(), key: "a", metaKey: true } as any;
    expect(IsCapsLockModified(event)).toBeNull();
});

it("returns null for space key", () => {
    const event = { altKey: false, ctrlKey: false, getModifierState: vi.fn(), key: " ", metaKey: false } as any;
    expect(IsCapsLockModified(event)).toBeNull();
});

it("returns null for safe key", () => {
    const event = { altKey: false, ctrlKey: false, getModifierState: vi.fn(), key: "1", metaKey: false } as any;
    expect(IsCapsLockModified(event)).toBeNull();
});

it("returns caps lock state for other keys", () => {
    const mockGetModifierState = vi.fn(() => true);
    const event = {
        altKey: false,
        ctrlKey: false,
        getModifierState: mockGetModifierState,
        key: "a",
        metaKey: false,
    } as any;
    expect(IsCapsLockModified(event)).toBe(true);
    expect(mockGetModifierState).toHaveBeenCalledWith("CapsLock");
});

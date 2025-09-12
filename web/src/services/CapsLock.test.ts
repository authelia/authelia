import { IsCapsLockModified } from "@services/CapsLock";

it("returns null for key length not 1", () => {
    const event = { key: "Enter", ctrlKey: false, altKey: false, metaKey: false, getModifierState: vi.fn() } as any;
    expect(IsCapsLockModified(event)).toBeNull();
});

it("returns null for ctrl key", () => {
    const event = { key: "a", ctrlKey: true, altKey: false, metaKey: false, getModifierState: vi.fn() } as any;
    expect(IsCapsLockModified(event)).toBeNull();
});

it("returns null for alt key", () => {
    const event = { key: "a", ctrlKey: false, altKey: true, metaKey: false, getModifierState: vi.fn() } as any;
    expect(IsCapsLockModified(event)).toBeNull();
});

it("returns null for meta key", () => {
    const event = { key: "a", ctrlKey: false, altKey: false, metaKey: true, getModifierState: vi.fn() } as any;
    expect(IsCapsLockModified(event)).toBeNull();
});

it("returns null for space key", () => {
    const event = { key: " ", ctrlKey: false, altKey: false, metaKey: false, getModifierState: vi.fn() } as any;
    expect(IsCapsLockModified(event)).toBeNull();
});

it("returns null for safe key", () => {
    const event = { key: "1", ctrlKey: false, altKey: false, metaKey: false, getModifierState: vi.fn() } as any;
    expect(IsCapsLockModified(event)).toBeNull();
});

it("returns caps lock state for other keys", () => {
    const mockGetModifierState = vi.fn(() => true);
    const event = {
        key: "a",
        ctrlKey: false,
        altKey: false,
        metaKey: false,
        getModifierState: mockGetModifierState,
    } as any;
    expect(IsCapsLockModified(event)).toBe(true);
    expect(mockGetModifierState).toHaveBeenCalledWith("CapsLock");
});

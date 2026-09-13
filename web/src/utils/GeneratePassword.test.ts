import { generateRandomPassword } from "@utils/GeneratePassword";

const LOWER = /[a-z]/;
const UPPER = /[A-Z]/;
const DIGIT = /[0-9]/;
const SPECIAL = /[!@#$%^&*_\-+=]/;
const ALLOWED = /^[a-zA-Z0-9!@#$%^&*_\-+=]+$/;

it("generates a password of the requested length", () => {
    expect(generateRandomPassword(16)).toHaveLength(16);
    expect(generateRandomPassword(4)).toHaveLength(4);
    expect(generateRandomPassword(64)).toHaveLength(64);
});

it("always contains at least one character of each class", () => {
    for (let i = 0; i < 50; i++) {
        const password = generateRandomPassword(8);

        expect(password).toMatch(LOWER);
        expect(password).toMatch(UPPER);
        expect(password).toMatch(DIGIT);
        expect(password).toMatch(SPECIAL);
    }
});

it("only uses characters from the allowed alphabet", () => {
    for (let i = 0; i < 20; i++) {
        expect(generateRandomPassword(32)).toMatch(ALLOWED);
    }
});

it("produces different passwords on successive calls", () => {
    const passwords = new Set(Array.from({ length: 10 }, () => generateRandomPassword(16)));

    expect(passwords.size).toBe(10);
});

it("uses the cryptographic random source", () => {
    const spy = vi.spyOn(crypto, "getRandomValues");

    generateRandomPassword(12);

    expect(spy).toHaveBeenCalled();

    spy.mockRestore();
});

import type { TestingLibraryMatchers } from "@testing-library/jest-dom/matchers";

declare module "vitest" {
    interface Matchers<
        R extends Promise<void> | void = Promise<void> | void,
        T = unknown,
    > extends TestingLibraryMatchers<T, R> {}
}
